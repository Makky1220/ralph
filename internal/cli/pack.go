package cli

import (
	"errors"
	"fmt"
	"io/fs"
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
			return addPack(".", args[0], force)
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

func addPack(targetDir, lang string, force bool) error {
	prr, err := renderPackInto(targetDir, lang, force)
	if err != nil {
		return err
	}

	printRenderSummary("pack/"+lang, prr.result)

	manifestPath := filepath.Join(targetDir, ".ralph", "manifest.toml")
	m, err := scaffold.ReadManifest(manifestPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Printf("\n  ⚠ No manifest found — run 'ralph init' first to track pack files.\n")
			return nil
		}
		return fmt.Errorf("reading manifest: %w", err)
	}

	if !containsPack(m.Meta.Packs, lang) {
		m.Meta.Packs = append(m.Meta.Packs, lang)
	}

	for path, hash := range prr.hashes {
		if baselinePath, ok := prr.baselinePaths[path]; ok {
			m.SetFileWithBaseline(path, hash, baselinePath)
		} else {
			m.SetFile(path, hash)
		}
	}

	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return fmt.Errorf("creating .ralph dir: %w", err)
	}
	if err := m.Write(manifestPath); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	fmt.Printf("  ✓ .ralph/manifest.toml (pack %s tracked)\n", lang)

	return nil
}

func containsPack(packs []string, lang string) bool {
	for _, p := range packs {
		if p == lang {
			return true
		}
	}
	return false
}
