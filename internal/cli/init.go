package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"

	"github.com/thomas0124/ralph/internal/scaffold"
)

func newInitCmd() *cobra.Command {
	var (
		nonInteractive bool
		force          bool
	)

	cmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialize a new project with Claude Code harness scaffold",
		Long: `Scaffolds a project with Claude Code configurations, hooks, skills,
agents, rules, and pipeline settings.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			}
			absDir, err := filepath.Abs(targetDir)
			if err != nil {
				return fmt.Errorf("resolving directory: %w", err)
			}
			if nonInteractive {
				return runInitNonInteractive(absDir, force)
			}
			return runInitInteractive(absDir, force)
		},
	}

	cmd.Flags().BoolVar(&nonInteractive, "yes", false, "skip interactive prompts, use defaults")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files without prompting")

	return cmd
}

type initConfig struct {
	ProjectName string
	Packs       []string
}

func runInitInteractive(targetDir string, force bool) error {
	manifestPath := filepath.Join(targetDir, ".ralph", "manifest.toml")
	if _, err := os.Stat(manifestPath); err == nil {
		fmt.Printf("\nExisting project detected. Running upgrade instead...\n\n")
		return runUpgrade(targetDir, false)
	}

	defaultName := filepath.Base(targetDir)

	availPacks, err := scaffold.AvailablePacks()
	if err != nil {
		return fmt.Errorf("listing packs: %w", err)
	}

	cfg := initConfig{
		ProjectName: defaultName,
		Packs:       availPacks,
	}

	packOptions := make([]huh.Option[string], len(availPacks))
	for i, p := range availPacks {
		packOptions[i] = huh.NewOption(p, p).Selected(true)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Value(&cfg.ProjectName),
			huh.NewMultiSelect[string]().
				Title("Language packs (optional)").
				Options(packOptions...).
				Value(&cfg.Packs),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("interactive form: %w", err)
	}

	return executeInit(targetDir, cfg, force)
}

func runInitNonInteractive(targetDir string, force bool) error {
	availPacks, err := scaffold.AvailablePacks()
	if err != nil {
		return fmt.Errorf("listing packs: %w", err)
	}

	cfg := initConfig{
		ProjectName: filepath.Base(targetDir),
		Packs:       availPacks,
	}

	return executeInit(targetDir, cfg, force)
}

func executeInit(targetDir string, cfg initConfig, force bool) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	manifestPath := filepath.Join(targetDir, ".ralph", "manifest.toml")
	if _, err := os.Stat(manifestPath); err == nil {
		fmt.Printf("\nExisting project detected. Running upgrade instead...\n\n")
		return runUpgrade(targetDir, false)
	}

	fmt.Printf("\nScaffolding %q into %s ...\n\n", cfg.ProjectName, targetDir)

	baseFS, err := scaffold.BaseFS()
	if err != nil {
		return fmt.Errorf("loading base templates: %w", err)
	}

	result, hashes, err := scaffold.RenderFS(baseFS, scaffold.RenderOptions{
		TargetDir: targetDir,
		Overwrite: force,
	})
	if err != nil {
		return fmt.Errorf("rendering base templates: %w", err)
	}
	baselinePaths, err := writeRenderedBaselines(targetDir, baseFS, "", result)
	if err != nil {
		return err
	}
	printRenderSummary("base", result)

	for _, pack := range cfg.Packs {
		prr, err := renderPackInto(targetDir, pack, force)
		if err != nil {
			fmt.Printf("  ⚠ pack %s: %v\n", pack, err)
			continue
		}
		for k, v := range prr.hashes {
			hashes[k] = v
		}
		for k, v := range prr.baselinePaths {
			baselinePaths[k] = v
		}
		printRenderSummary("pack/"+pack, prr.result)
	}

	manifest := scaffold.NewManifest(Version)
	manifest.Meta.Packs = cfg.Packs
	for path, hash := range hashes {
		if baselinePath, ok := baselinePaths[path]; ok {
			manifest.SetFileWithBaseline(path, hash, baselinePath)
			continue
		}
		manifest.SetFile(path, hash)
	}

	manifestDir := filepath.Join(targetDir, ".ralph")
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("creating .ralph dir: %w", err)
	}
	mPath := filepath.Join(manifestDir, "manifest.toml")
	if err := manifest.Write(mPath); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	fmt.Printf("  ✓ .ralph/manifest.toml\n")

	gitDir := filepath.Join(targetDir, ".git")
	if _, err := os.Stat(gitDir); errors.Is(err, fs.ErrNotExist) {
		if gitBin, err := exec.LookPath("git"); err == nil {
			cmd := exec.Command(gitBin, "init")
			cmd.Dir = targetDir
			if out, err := cmd.CombinedOutput(); err != nil {
				fmt.Printf("  ⚠ git init failed: %s\n", out)
			} else {
				fmt.Printf("  ✓ git init\n")
			}
		}
	} else {
		fmt.Printf("  ✓ .git exists (skipped)\n")
	}

	installManagedGitHooks(targetDir, os.Stdout, os.Stderr)

	fmt.Printf("\nDone. Next steps:\n")
	if targetDir != "." {
		fmt.Printf("  cd %s\n", targetDir)
	}
	fmt.Printf("  Edit CLAUDE.md to describe your project\n")
	fmt.Printf("  ralph doctor   — verify setup\n")

	return nil
}

func writeRenderedBaselines(targetDir string, src fs.FS, prefix string, result *scaffold.RenderResult) (map[string]string, error) {
	out := make(map[string]string)
	written := make(map[string]bool, len(result.Created)+len(result.Overwritten))
	for _, path := range result.Created {
		written[path] = true
	}
	for _, path := range result.Overwritten {
		written[path] = true
	}
	for path := range written {
		content, err := fs.ReadFile(src, path)
		if err != nil {
			return nil, fmt.Errorf("reading baseline source %s: %w", path, err)
		}
		manifestPath := filepath.Join(prefix, path)
		baselinePath, err := scaffold.WriteBaseline(targetDir, manifestPath, content)
		if err != nil {
			return nil, err
		}
		out[manifestPath] = baselinePath
	}
	return out, nil
}

func printRenderSummary(label string, result *scaffold.RenderResult) {
	created := len(result.Created)
	overwritten := len(result.Overwritten)
	skipped := len(result.Skipped)
	total := created + overwritten + skipped
	fmt.Printf("  ✓ %s (%d files: %d created, %d updated, %d skipped)\n",
		label, total, created, overwritten, skipped)
}
