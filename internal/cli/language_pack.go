package cli

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/thomas0124/ralph/internal/scaffold"
)

// packRenderResult bundles the output of renderPackInto.
type packRenderResult struct {
	result        *scaffold.RenderResult
	hashes        map[string]string
	baselinePaths map[string]string
}

// renderPackInto renders a language pack into packs/languages/<lang>/ inside
// targetDir, maps rule.md to .claude/rules/ralph/<lang>.md, and writes
// baselines so the upgrade engine can diff pack files later.
func renderPackInto(targetDir, lang string, force bool) (*packRenderResult, error) {
	packFS, err := scaffold.PackFS(lang)
	if err != nil {
		return nil, fmt.Errorf("language pack %q not found: %w", lang, err)
	}

	packDir := filepath.Join(targetDir, packRelDir(lang))
	packResult, packHashes, err := scaffold.RenderFS(packFS, scaffold.RenderOptions{
		TargetDir: packDir,
		Overwrite: force,
		SkipPaths: packRenderSkipPaths,
	})
	if err != nil {
		return nil, fmt.Errorf("rendering pack %s: %w", lang, err)
	}

	out := &packRenderResult{
		result:        packResult,
		hashes:        make(map[string]string),
		baselinePaths: make(map[string]string),
	}

	packPrefix := packRelDir(lang)
	for k, v := range packHashes {
		out.hashes[filepath.Join(packPrefix, k)] = v
	}

	bls, err := writeRenderedBaselines(targetDir, packFS, packPrefix, packResult)
	if err != nil {
		return nil, err
	}
	for k, v := range bls {
		out.baselinePaths[k] = v
	}

	ruleContent, ok, err := packRuleContent(packFS)
	if err != nil {
		return nil, fmt.Errorf("reading pack %s rule: %w", lang, err)
	}
	if ok {
		rulePath := packRuleRelPath(lang)
		ruleResult, ruleHash, err := renderMappedFile(targetDir, rulePath, ruleContent, force)
		if err != nil {
			return nil, fmt.Errorf("rendering pack %s rule: %w", lang, err)
		}
		out.hashes[rulePath] = ruleHash
		if len(ruleResult.Created)+len(ruleResult.Overwritten) > 0 {
			baselinePath, err := scaffold.WriteBaseline(targetDir, rulePath, ruleContent)
			if err != nil {
				return nil, err
			}
			out.baselinePaths[rulePath] = baselinePath
		}
		mergeRenderResult(packResult, ruleResult)
	}

	return out, nil
}

const packRuleSourcePath = "rule.md"

var packRenderSkipPaths = map[string]bool{
	packRuleSourcePath: true,
}

func packRelDir(pack string) string {
	return filepath.Join("packs", "languages", pack)
}

// packRuleRelPath returns the manifest key for a pack's rule file.
// Uses path.Join (not filepath.Join) because manifest keys use slash separators.
func packRuleRelPath(pack string) string {
	return path.Join(".claude", "rules", "ralph", pack+".md")
}

func isInsideDir(absDir, absPath string) bool {
	return absPath == absDir || strings.HasPrefix(absPath, absDir+string(filepath.Separator))
}
