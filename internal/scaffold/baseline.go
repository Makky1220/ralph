package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BaselinePath(relPath string) (string, error) {
	safe := strings.ReplaceAll(filepath.ToSlash(relPath), "/", "__")
	return filepath.Join(".ralph", "baselines", safe), nil
}

func WriteBaseline(targetDir, relPath string, content []byte) (string, error) {
	basePath, err := BaselinePath(relPath)
	if err != nil {
		return "", err
	}
	abs := filepath.Join(targetDir, basePath)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return "", fmt.Errorf("creating baseline dir: %w", err)
	}
	if err := os.WriteFile(abs, content, 0644); err != nil {
		return "", fmt.Errorf("writing baseline %s: %w", basePath, err)
	}
	return basePath, nil
}

func ReadBaseline(targetDir string, mf ManifestFile) ([]byte, error) {
	if mf.BaselinePath == "" {
		return nil, fmt.Errorf("no baseline path recorded")
	}
	abs := filepath.Join(targetDir, mf.BaselinePath)
	return os.ReadFile(abs)
}
