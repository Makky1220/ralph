package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/thomas0124/ralph/internal/scaffold"
	"github.com/thomas0124/ralph/internal/upgrade"
)

func newUpgradeCmd() *cobra.Command {
	var (
		force       bool
		dryRun      bool
		diffPreview bool
		pager       string
	)

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Update scaffold files to the latest template version",
		Long: `Compares the current project files against the embedded templates,
auto-updates unchanged files, and prompts for conflict resolution on edited files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if diffPreview {
				dryRun = true
			}
			return runUpgradeWithOptions(".", upgradeOptions{
				Force:       force,
				DryRun:      dryRun,
				DiffPreview: diffPreview,
				Pager:       pager,
			})
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "overwrite all files without prompting")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview upgrade actions without writing files")
	cmd.Flags().BoolVar(&diffPreview, "diff", false, "show conflict diffs without writing files (implies --dry-run)")
	cmd.Flags().StringVar(&pager, "pager", pagerAuto, "dry-run diff pager mode: auto, always, or never")

	return cmd
}

const (
	pagerAuto   = "auto"
	pagerAlways = "always"
	pagerNever  = "never"
)

type upgradeOptions struct {
	Force       bool
	DryRun      bool
	DiffPreview bool
	Pager       string
}

func runUpgrade(targetDir string, force bool) error {
	return runUpgradeWithOptions(targetDir, upgradeOptions{
		Force: force,
		Pager: pagerAuto,
	})
}

func runUpgradeWithOptions(targetDir string, opts upgradeOptions) error {
	return runUpgradeIOWithOptions(targetDir, opts, os.Stdin, os.Stdout, os.Stderr, shouldColorize(os.Stdout))
}

func shouldColorize(out *os.File) bool {
	if v := os.Getenv("NO_COLOR"); v != "" {
		return false
	}
	if out == nil {
		return false
	}
	return term.IsTerminal(out.Fd())
}

func validatePagerMode(mode string) error {
	switch mode {
	case pagerAuto, pagerAlways, pagerNever:
		return nil
	default:
		return fmt.Errorf("invalid --pager %q (want auto, always, or never)", mode)
	}
}

func writeDiffOutput(diff string, out, errOut io.Writer, pagerMode string) {
	if !shouldUsePager(pagerMode, out) {
		_, _ = io.WriteString(out, diff)
		return
	}
	if err := writeThroughPager(diff, out, errOut); err != nil {
		writef(errOut, "    (pager failed: %v — writing diff directly)\n", err)
		_, _ = io.WriteString(out, diff)
	}
}

func runUpgradeIOWithOptions(targetDir string, opts upgradeOptions, in io.Reader, out, errOut io.Writer, colorize bool) error {
	if opts.Pager == "" {
		opts.Pager = pagerAuto
	}
	if opts.DiffPreview {
		opts.DryRun = true
	}

	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("resolving directory: %w", err)
	}

	manifestPath := filepath.Join(absDir, ".ralph", "manifest.toml")
	if _, err := os.Stat(manifestPath); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("no .ralph/manifest.toml found — run 'ralph init' first")
	}

	oldManifest, err := scaffold.ReadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("reading manifest: %w", err)
	}

	writef(out, "Checking for updates...\n")
	writef(out, "  Current: %s → Available: %s\n\n", oldManifest.Meta.Version, Version)

	baseFS, err := scaffold.BaseFS()
	if err != nil {
		return fmt.Errorf("loading templates: %w", err)
	}

	baseManifest := splitManifestForBase(oldManifest)
	diffs, err := upgrade.ComputeDiffsWithManifest(baseManifest, absDir, baseFS, true)
	if err != nil {
		return fmt.Errorf("computing diffs: %w", err)
	}

	installedPacks := oldManifest.Meta.Packs
	availablePacks, apErr := scaffold.AvailablePacks()

	preservedPackEntries := make(map[string]scaffold.ManifestFile)
	retainedPacks := make([]string, 0, len(installedPacks))

	if apErr != nil {
		writef(errOut, "Warning: unable to list available packs: %v\n", apErr)
		for _, pack := range installedPacks {
			preservePackState(oldManifest, pack, preservedPackEntries)
			retainedPacks = append(retainedPacks, pack)
		}
		installedPacks = nil
	}
	available := make(map[string]bool, len(availablePacks))
	for _, p := range availablePacks {
		available[p] = true
	}

	for _, pack := range installedPacks {
		if !available[pack] {
			writef(errOut, "Notice: pack %q no longer in templates — manifest tracking dropped\n", pack)
			continue
		}

		packFS, pErr := scaffold.PackFS(pack)
		if pErr != nil {
			writef(errOut, "Warning: pack %s load failed: %v\n", pack, pErr)
			preservePackState(oldManifest, pack, preservedPackEntries)
			retainedPacks = append(retainedPacks, pack)
			continue
		}

		packManifest := splitManifestForPackOnly(oldManifest, pack)
		packDir := filepath.Join(absDir, packRelDir(pack))
		packPayloadDiffs, pDiffErr := upgrade.ComputeDiffsWithManifest(packManifest, packDir, packFS, false)
		if pDiffErr == nil {
			for i := range packPayloadDiffs {
				packPayloadDiffs[i].Path = filepath.Join(packRelDir(pack), packPayloadDiffs[i].Path)
			}
			diffs = append(diffs, packPayloadDiffs...)
		}
		if content, ok, err := packRuleContent(packFS); err == nil && ok {
			diffs = append(diffs, upgrade.ComputeFileDiff(oldManifest, absDir, packRuleRelPath(pack), content))
		}
		retainedPacks = append(retainedPacks, pack)
	}

	if opts.DryRun {
		renderUpgradePreview(diffs, absDir, Version, out, errOut, colorize, opts)
		return nil
	}

	manifest := scaffold.NewManifest(Version)
	manifest.Meta.Packs = retainedPacks
	maps.Copy(manifest.Files, preservedPackEntries)

	reader := bufio.NewReader(in)
	applyPlan := &upgradeApplyPlan{}

	for _, d := range diffs {
		switch d.Action {
		case upgrade.ActionAutoUpdate:
			if err := applyPlan.addTemplateWrite(manifest, d.Path, d.NewHash, d.NewContent,
				fmt.Sprintf("  ✓ %s (unchanged, auto-update)\n", d.Path)); err != nil {
				return err
			}

		case upgrade.ActionConflict:
			if opts.Force {
				if err := applyPlan.addTemplateWrite(manifest, d.Path, d.NewHash, d.NewContent,
					fmt.Sprintf("  ✓ %s (force overwritten)\n", d.Path)); err != nil {
					return err
				}
				continue
			}
			result := resolveConflict(d, oldManifest, absDir, Version, reader, out, errOut, colorize)
			applyPlan.needsConfirmation = true
			switch result.kind {
			case resolutionOverwrite:
				if err := applyPlan.addTemplateWrite(manifest, d.Path, d.NewHash, d.NewContent, ""); err != nil {
					return err
				}
			case resolutionSkip:
				hash := d.DiskHash
				if hash == "" {
					if d.OldHash != "" {
						hash = d.OldHash
					} else {
						hash = d.NewHash
					}
				}
				manifest.SetFileUnmanaged(d.Path, hash)
				applyPlan.addMessage(fmt.Sprintf("  ⊘ %s (kept local)\n", d.Path))
				applyPlan.skipped++
			case resolutionResolved:
				if err := applyPlan.addResolvedWrite(manifest, d.Path, d.NewHash, d.NewContent, result.content); err != nil {
					return err
				}
			}

		case upgrade.ActionAdd:
			if err := applyPlan.addTemplateWrite(manifest, d.Path, d.NewHash, d.NewContent,
				fmt.Sprintf("  + %s (new file)\n", d.Path)); err != nil {
				return err
			}

		case upgrade.ActionRemove:
			applyPlan.addMessage(fmt.Sprintf("  ⚠ %s (removed from template — review and delete manually)\n", d.Path))
			applyPlan.notified++

		case upgrade.ActionSkip:
			prev, hadEntry := oldManifest.Files[d.Path]
			wasUnmanaged := hadEntry && !prev.Managed
			switch {
			case opts.Force && wasUnmanaged && d.NewContent != nil:
				if err := applyPlan.addTemplateWrite(manifest, d.Path, d.NewHash, d.NewContent,
					fmt.Sprintf("  ✓ %s (force re-adopted)\n", d.Path)); err != nil {
					return err
				}
			case wasUnmanaged:
				if prev.State == "" {
					prev.State = scaffold.FileStateUnmanaged
				}
				manifest.Files[d.Path] = prev
			default:
				if prev, ok := oldManifest.Files[d.Path]; ok {
					prev = prev.WithTemplateHash(d.NewHash)
					manifest.Files[d.Path] = prev
				} else {
					manifest.SetFile(d.Path, d.NewHash)
				}
			}
		}
	}

	if applyPlan.needsConfirmation {
		writef(out, "\nApply these changes? [y/N] ")
		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			writef(errOut, "\n  (non-interactive input detected, no changes applied)\n")
			return nil
		}
		choice := strings.ToLower(strings.TrimSpace(line))
		if choice != "y" && choice != "yes" {
			writef(out, "\nNo changes applied.\n")
			return nil
		}
	}

	if err := applyPlan.apply(absDir); err != nil {
		return err
	}
	applyPlan.writeMessages(out)

	if err := manifest.Write(manifestPath); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	writef(out, "\n  Updated: %d files\n", applyPlan.updated)
	writef(out, "  Skipped: %d files (user-modified)\n", applyPlan.skipped)
	if applyPlan.notified > 0 {
		writef(out, "  Removed from template: %d files (review manually)\n", applyPlan.notified)
	}
	writef(out, "  Manifest updated: .ralph/manifest.toml\n")

	return nil
}

// splitManifestForPackOnly builds a sub-manifest containing only the payload
// files for a single pack (i.e. files inside packs/languages/<pack>/), with
// keys stripped of the packRelDir prefix so they match what ComputeDiffsWithManifest
// expects when given a PackFS rooted at the pack directory.
func splitManifestForPackOnly(src *scaffold.Manifest, pack string) *scaffold.Manifest {
	out := scaffold.NewManifest(src.Meta.Version)
	prefix := filepath.ToSlash(packRelDir(pack)) + "/"
	for k, v := range src.Files {
		slashPath := filepath.ToSlash(k)
		if strings.HasPrefix(slashPath, prefix) {
			stripped := strings.TrimPrefix(slashPath, prefix)
			out.Files[stripped] = v
		}
	}
	return out
}

func isPackManagedPath(pack, slashPath string) bool {
	packPrefix := filepath.ToSlash(packRelDir(pack)) + "/"
	ruleFile := filepath.ToSlash(packRuleRelPath(pack))
	return strings.HasPrefix(slashPath, packPrefix) || slashPath == ruleFile
}

func splitManifestForBase(m *scaffold.Manifest) *scaffold.Manifest {
	out := scaffold.NewManifest(m.Meta.Version)
	out.Meta = m.Meta
	out.Files = make(map[string]scaffold.ManifestFile, len(m.Files))
	for k, v := range m.Files {
		slashPath := filepath.ToSlash(k)
		managed := false
		for _, pack := range m.Meta.Packs {
			if isPackManagedPath(pack, slashPath) {
				managed = true
				break
			}
		}
		if !managed {
			out.Files[k] = v
		}
	}
	return out
}

func preservePackState(src *scaffold.Manifest, pack string, dst map[string]scaffold.ManifestFile) {
	for k, v := range src.Files {
		if isPackManagedPath(pack, filepath.ToSlash(k)) {
			dst[k] = v
		}
	}
}

type plannedFileWrite struct {
	relPath string
	content []byte
}

type upgradeApplyPlan struct {
	writes            []plannedFileWrite
	baselines         []plannedFileWrite
	messages          []string
	updated           int
	skipped           int
	notified          int
	needsConfirmation bool
}

func (p *upgradeApplyPlan) addMessage(msg string) {
	if msg != "" {
		p.messages = append(p.messages, msg)
	}
}

func (p *upgradeApplyPlan) addTemplateWrite(manifest *scaffold.Manifest, relPath, hash string, content []byte, message string) error {
	return p.addResolvedWrite(manifest, relPath, hash, content, content)
}

func (p *upgradeApplyPlan) addResolvedWrite(manifest *scaffold.Manifest, relPath, templateHash string, templateContent, resolvedContent []byte) error {
	baselinePath, err := scaffold.BaselinePath(relPath)
	if err != nil {
		return err
	}
	diskHash := scaffold.HashBytes(resolvedContent)
	state := scaffold.FileStatePartial
	if diskHash == templateHash {
		state = scaffold.FileStateManaged
	}
	manifest.SetFileResolvedWithBaseline(relPath, templateHash, diskHash, state, baselinePath)
	p.writes = append(p.writes, plannedFileWrite{relPath: relPath, content: resolvedContent})
	p.baselines = append(p.baselines, plannedFileWrite{relPath: relPath, content: templateContent})
	p.updated++
	return nil
}

func (p *upgradeApplyPlan) apply(absDir string) error {
	for _, w := range p.writes {
		target := filepath.Join(absDir, filepath.FromSlash(w.relPath))
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("creating parent for %s: %w", w.relPath, err)
		}
		if err := os.WriteFile(target, w.content, scaffold.FilePerm(w.relPath)); err != nil {
			return fmt.Errorf("writing %s: %w", w.relPath, err)
		}
	}
	for _, b := range p.baselines {
		if _, err := scaffold.WriteBaseline(absDir, b.relPath, b.content); err != nil {
			return err
		}
	}
	return nil
}

func (p *upgradeApplyPlan) writeMessages(out io.Writer) {
	for _, msg := range p.messages {
		writef(out, "%s", msg)
	}
}

type resolution int

const (
	resolutionSkip resolution = iota
	resolutionOverwrite
	resolutionResolved
)

type conflictResult struct {
	kind    resolution
	content []byte
}

func resolveConflict(d upgrade.FileDiff, oldManifest *scaffold.Manifest, absDir, version string, in *bufio.Reader, out, errOut io.Writer, colorize bool) conflictResult {
	writef(out, "  ⚠ %s (modified locally)\n", d.Path)

	prev, hasPrev := oldManifest.Files[d.Path]
	if hasPrev && prev.IsBaselineAvailable() {
		baselineBytes, err := scaffold.ReadBaseline(absDir, prev)
		if err == nil {
			localBytes, err := os.ReadFile(filepath.Join(absDir, d.Path))
			if err == nil {
				mergePlan := upgrade.PlanMerge(baselineBytes, localBytes, d.NewContent)
				if mergePlan.ConflictCount() == 0 {
					return conflictResult{kind: resolutionResolved, content: d.NewContent}
				}
			}
		}
	}

	showDiff(d, absDir, version, out, errOut, colorize)
	for {
		writef(out, "    [o]verwrite / [s]kip / [d]iff ? ")
		line, err := in.ReadString('\n')
		if err != nil && line == "" {
			writef(errOut, "\n  (non-interactive input detected, skipping)\n")
			return conflictResult{kind: resolutionSkip}
		}
		switch strings.TrimSpace(line) {
		case "o", "overwrite":
			return conflictResult{kind: resolutionOverwrite}
		case "s", "skip":
			return conflictResult{kind: resolutionSkip}
		case "d", "diff":
			showDiff(d, absDir, version, out, errOut, colorize)
		}
	}
}

func showDiff(d upgrade.FileDiff, absDir, version string, out, errOut io.Writer, colorize bool) {
	localPath := filepath.Join(absDir, d.Path)
	localBytes, err := os.ReadFile(localPath)
	if err != nil {
		writef(errOut, "    (could not read %s: %v)\n", d.Path, err)
		return
	}
	diff := upgrade.UnifiedDiff(localBytes, d.NewContent, "local", fmt.Sprintf("template (%s)", version))
	diff = omitRangeHeaders(diff)
	if diff == "" {
		writef(out, "    (no textual difference — manifest hash drift only)\n")
	} else {
		if colorize {
			diff = upgrade.Colorize(diff)
		}
		_, _ = io.WriteString(out, diff)
	}
}

func omitRangeHeaders(diff string) string {
	if diff == "" {
		return ""
	}
	var b strings.Builder
	for _, line := range strings.SplitAfter(diff, "\n") {
		if strings.HasPrefix(strings.TrimSuffix(line, "\n"), "@@ ") {
			continue
		}
		b.WriteString(line)
	}
	return b.String()
}

func renderUpgradePreview(diffs []upgrade.FileDiff, absDir, version string, out, errOut io.Writer, colorize bool, opts upgradeOptions) {
	var autoUpdate, conflict, add, remove, skip int
	for _, d := range diffs {
		switch d.Action {
		case upgrade.ActionAutoUpdate:
			autoUpdate++
		case upgrade.ActionConflict:
			conflict++
		case upgrade.ActionAdd:
			add++
		case upgrade.ActionRemove:
			remove++
		case upgrade.ActionSkip:
			skip++
		}
	}

	writef(out, "Upgrade preview (dry run)\n")
	writef(out, "  auto-update: %d files\n", autoUpdate)
	writef(out, "  conflict:    %d files\n", conflict)
	writef(out, "  add:         %d files\n", add)
	writef(out, "  remove:      %d files\n", remove)
	writef(out, "  skip:        %d files\n", skip)

	if len(diffs) == 0 {
		writef(out, "\nNo changes.\n")
		return
	}

	writef(out, "\nFiles:\n")
	for _, d := range diffs {
		label := actionLabel(d.Action)
		writef(out, "  %-11s %s\n", label, d.Path)
	}

	if opts.DiffPreview {
		for _, d := range diffs {
			if d.Action != upgrade.ActionConflict && d.Action != upgrade.ActionAutoUpdate {
				continue
			}
			writef(out, "\n--- %s ---\n", d.Path)
			showDiff(d, absDir, version, out, errOut, colorize)
		}
	}
}

func actionLabel(action upgrade.FileAction) string {
	switch action {
	case upgrade.ActionAutoUpdate:
		return "auto-update"
	case upgrade.ActionConflict:
		return "conflict"
	case upgrade.ActionAdd:
		return "add"
	case upgrade.ActionRemove:
		return "remove"
	case upgrade.ActionSkip:
		return "skip"
	default:
		return "unknown"
	}
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func shouldUsePager(pagerMode string, out io.Writer) bool {
	switch pagerMode {
	case pagerNever:
		return false
	case pagerAlways:
		return true
	case pagerAuto:
		f, ok := out.(*os.File)
		return ok && term.IsTerminal(f.Fd())
	default:
		return false
	}
}

func writeThroughPager(diff string, out, errOut io.Writer) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less -R"
	}
	parts := strings.Fields(pager)
	if len(parts) == 0 {
		return fmt.Errorf("empty pager")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = strings.NewReader(diff)
	cmd.Stdout = out
	cmd.Stderr = errOut
	return cmd.Run()
}

