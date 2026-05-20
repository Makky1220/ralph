package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thomas0124/ralph/internal/scaffold"
)

func newPackCmd() *cobra.Command {
	pack := &cobra.Command{
		Use:   "pack",
		Short: "Manage language packs",
	}

	pack.AddCommand(newPackAddCmd())
	pack.AddCommand(newPackListCmd())

	return pack
}

func newPackAddCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "add <language>",
		Short: "Add a language pack to the project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPackAdd(".", args[0], force)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing pack files")
	return cmd
}

func newPackListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available language packs",
		RunE: func(cmd *cobra.Command, args []string) error {
			packs, err := scaffold.AvailablePacks()
			if err != nil {
				return err
			}
			if len(packs) == 0 {
				fmt.Println("No language packs available.")
				return nil
			}
			fmt.Println("Available language packs:")
			for _, p := range packs {
				fmt.Printf("  %s\n", p)
			}
			return nil
		},
	}
}

func runPackAdd(targetDir, lang string, force bool) error {
	packFS, err := scaffold.PackFS(lang)
	if err != nil {
		return fmt.Errorf("loading pack %s: %w", lang, err)
	}

	packDir := filepath.Join(targetDir, "packs", "languages", lang)
	result, hashes, err := scaffold.RenderFS(packFS, scaffold.RenderOptions{
		TargetDir: packDir,
		Overwrite: force,
		SkipPaths: map[string]bool{"rule.md": true},
	})
	if err != nil {
		return fmt.Errorf("rendering pack: %w", err)
	}

	_ = hashes
	_ = result

	ruleContent, ruleOK, err := packRuleContent(packFS)
	if err != nil {
		return fmt.Errorf("reading pack rule: %w", err)
	}
	if ruleOK {
		rulePath := filepath.Join(targetDir, ".claude", "rules", lang+".md")
		if _, statErr := os.Stat(rulePath); statErr == nil && !force {
			fmt.Printf("  ⊘ .claude/rules/%s.md (exists, skipped)\n", lang)
		} else {
			if err := os.MkdirAll(filepath.Dir(rulePath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(rulePath, ruleContent, 0644); err != nil {
				return err
			}
			fmt.Printf("  ✓ .claude/rules/%s.md\n", lang)
		}
	}

	created := len(result.Created)
	overwritten := len(result.Overwritten)
	skipped := len(result.Skipped)
	fmt.Printf("\nPack %s: %d files created, %d updated, %d skipped\n", lang, created, overwritten, skipped)
	fmt.Printf("  Add packs/languages/%s/verify.sh to your verification pipeline.\n", lang)
	return nil
}
