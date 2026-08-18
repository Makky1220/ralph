package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"

	"github.com/thomas0124/ralph/internal/scaffold"
	"github.com/thomas0124/ralph/internal/upgrade"
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

// ownerForScaffoldPath returns the manifest v3 ownership attribute for a
// scaffolded file, keyed by its manifest-relative path.
func ownerForScaffoldPath(relPath string) string {
	relPath = filepath.ToSlash(relPath)
	switch relPath {
	case "AGENTS.md", ".gitignore":
		return scaffold.OwnerBlock
	case "CLAUDE.md", "ralph.toml", ".github/workflows/verify.yml", ".codex/AGENTS.override.md":
		return scaffold.OwnerSeed
	}
	if strings.HasPrefix(relPath, "docs/") {
		return scaffold.OwnerSeed
	}
	if strings.HasPrefix(relPath, ".ralph/local/") {
		return scaffold.OwnerSeed
	}
	return scaffold.OwnerCore
}

// blockSurface pairs a block-owned manifest path with the ralph surface
// token and marker style used to locate its managed block.
type blockSurface struct {
	path    string
	surface string
	style   upgrade.BlockMarkerStyle
}

// blockSurfaces lists every file whose ownership attribute is "block":
// files that are user-owned outside a single ralph-managed marker pair.
var blockSurfaces = []blockSurface{
	{path: "AGENTS.md", surface: "agents-md", style: upgrade.BlockMarkerHTML},
	{path: ".gitignore", surface: "gitignore", style: upgrade.BlockMarkerHash},
}

// reconcileBlockSurfaces appends the ralph managed block into pre-existing
// AGENTS.md / .gitignore files that RenderFS skipped because they already
// existed.
func reconcileBlockSurfaces(targetDir string, baseFS fs.FS, skipped []string, w io.Writer) (map[string]string, error) {
	skippedSet := make(map[string]bool, len(skipped))
	for _, p := range skipped {
		skippedSet[p] = true
	}

	diskHashes := make(map[string]string)
	for _, bs := range blockSurfaces {
		if !skippedSet[bs.path] {
			continue
		}

		templateContent, err := fs.ReadFile(baseFS, bs.path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("reading template %s: %w", bs.path, err)
		}
		interior, err := extractBlockInterior(templateContent, bs.surface, bs.style)
		if err != nil {
			return nil, fmt.Errorf("extracting managed block from template %s: %w", bs.path, err)
		}

		diskPath := filepath.Join(targetDir, bs.path)

		info, err := os.Lstat(diskPath)
		if err != nil {
			return nil, fmt.Errorf("reading existing %s: %w", bs.path, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			writef(w, "  ⚠ %s: is a symlink; left untouched\n", bs.path)
			continue
		}
		if !info.Mode().IsRegular() {
			writef(w, "  ⚠ %s: is not a regular file; left untouched\n", bs.path)
			continue
		}

		current, err := os.ReadFile(diskPath)
		if err != nil {
			return nil, fmt.Errorf("reading existing %s: %w", bs.path, err)
		}

		result := upgrade.UpdateManagedBlockStyled(current, bs.surface, interior, bs.style)
		switch result.Outcome {
		case upgrade.BlockAppended:
			if err := os.WriteFile(diskPath, result.Content, scaffold.FilePerm(bs.path)); err != nil {
				return nil, fmt.Errorf("writing %s: %w", bs.path, err)
			}
			diskHashes[bs.path] = scaffold.HashBytes(result.Content)
			writef(w, "  ✓ %s (appended ralph managed block)\n", bs.path)
		case upgrade.BlockMalformed:
			writef(w, "  ⚠ %s: existing ralph managed block is malformed (%s); left untouched\n", bs.path, result.Reason)
		default:
			// BlockUpdated / BlockUnchanged: a well-formed block already
			// exists on disk. Init only appends when a block is absent.
		}
	}
	return diskHashes, nil
}

// extractBlockInterior returns the bytes strictly between the BEGIN and END
// marker lines of a rendered template's managed block.
func extractBlockInterior(templateContent []byte, surface string, style upgrade.BlockMarkerStyle) ([]byte, error) {
	begin := upgrade.BeginMarkerStyled(surface, style)
	end := upgrade.EndMarkerStyled(style)

	lines := strings.Split(string(templateContent), "\n")
	beginIdx, endIdx := -1, -1
	for i, l := range lines {
		trimmed := strings.TrimSuffix(l, "\r")
		if trimmed == begin && beginIdx == -1 {
			beginIdx = i
		}
		if trimmed == end && beginIdx != -1 && endIdx == -1 {
			endIdx = i
		}
	}
	if beginIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("template block markers not found for surface %q", surface)
	}

	interior := lines[beginIdx+1 : endIdx]
	if len(interior) == 0 {
		return nil, nil
	}
	return []byte(strings.Join(interior, "\n") + "\n"), nil
}
