package upgrade

import (
	"errors"
	"fmt"
	"io/fs"
)

// AdvisoryDiff is the rendered outcome of comparing one AdvisoryEntry's disk
// content against the new template content.
type AdvisoryDiff struct {
	Path  string
	Owner string
	// Diff is the unified diff (disk vs template), labeled "local" and
	// "template (<version>)". Empty when contents are byte-identical, or
	// when Skipped is true.
	Diff string
	// Skipped is true when the template no longer has this path — nothing
	// to compare against, so no diff is rendered.
	Skipped bool
	// Note explains why Skipped is true. Empty otherwise.
	Note string
}

// RenderAdvisoryDiffs loads and renders unified diffs for each AdvisoryEntry
// in entries, comparing disk content in targetDir against the new template
// content in templateFS, labeled with templateVersion.
func RenderAdvisoryDiffs(targetDir string, templateFS fs.FS, templateVersion string, entries []AdvisoryEntry) ([]AdvisoryDiff, error) {
	diffs := make([]AdvisoryDiff, 0, len(entries))
	for _, entry := range entries {
		clean, err := cleanPathKey(entry.Path)
		if err != nil {
			return nil, fmt.Errorf("advisory path %q: %w", entry.Path, err)
		}

		localContent, _, err := readDiskFile(targetDir, clean)
		if err != nil {
			return nil, fmt.Errorf("reading disk file %q: %w", clean, err)
		}

		tmplContent, hasTemplate, err := readTemplateFile(templateFS, clean)
		if err != nil {
			return nil, fmt.Errorf("reading template file %q: %w", clean, err)
		}

		if !hasTemplate {
			diffs = append(diffs, AdvisoryDiff{
				Path:    clean,
				Owner:   entry.Owner,
				Skipped: true,
				Note:    "template no longer has this path",
			})
			continue
		}

		newLabel := fmt.Sprintf("template (%s)", templateVersion)
		diffs = append(diffs, AdvisoryDiff{
			Path:  clean,
			Owner: entry.Owner,
			Diff:  UnifiedDiff(localContent, tmplContent, "local", newLabel),
		})
	}
	return diffs, nil
}

// readTemplateFile reads path from templateFS, reporting hasTemplate=false
// (with no error) when the path does not exist.
func readTemplateFile(templateFS fs.FS, path string) (content []byte, hasTemplate bool, err error) {
	data, err := fs.ReadFile(templateFS, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return data, true, nil
}
