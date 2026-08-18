package cli

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/thomas0124/ralph/internal/scaffold"
)

func packRuleContent(packFS fs.FS) ([]byte, bool, error) {
	content, err := fs.ReadFile(packFS, "rule.md")
	if err != nil {
		return nil, false, nil
	}
	return content, true, nil
}

func packVerifyRelPath(lang string) string {
	return filepath.Join("packs", "languages", lang, "verify.sh")
}

func packVerifyContent(packFS fs.FS) ([]byte, bool, error) {
	content, err := fs.ReadFile(packFS, "verify.sh")
	if err != nil {
		return nil, false, nil
	}
	return content, true, nil
}

func mergeRenderResult(dst, src *scaffold.RenderResult) {
	dst.Created = append(dst.Created, src.Created...)
	dst.Overwritten = append(dst.Overwritten, src.Overwritten...)
	dst.Skipped = append(dst.Skipped, src.Skipped...)
}

func renderMappedFile(targetDir, relPath string, content []byte, force bool) (*scaffold.RenderResult, string, error) {
	result := &scaffold.RenderResult{}
	hash := scaffold.HashBytes(content)

	abs := filepath.Join(targetDir, relPath)
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, "", err
	}
	absAbs, err := filepath.Abs(abs)
	if err != nil {
		return nil, "", err
	}
	if !isInsideDir(absTargetDir, absAbs) {
		return nil, "", nil
	}

	_, statErr := os.Lstat(abs)
	exists := statErr == nil

	if exists && !force {
		result.Skipped = append(result.Skipped, relPath)
		return result, hash, nil
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return nil, "", err
	}
	if err := os.WriteFile(abs, content, scaffold.FilePerm(relPath)); err != nil {
		return nil, "", err
	}
	if exists {
		result.Overwritten = append(result.Overwritten, relPath)
	} else {
		result.Created = append(result.Created, relPath)
	}
	return result, hash, nil
}
