package scaffold

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type RenderOptions struct {
	TargetDir string
	Overwrite bool
	SkipPaths map[string]bool
}

type RenderResult struct {
	Created     []string
	Overwritten []string
	Skipped     []string
}

func RenderFS(src fs.FS, opts RenderOptions) (*RenderResult, map[string]string, error) {
	result := &RenderResult{}
	hashes := make(map[string]string)

	absTarget, err := filepath.Abs(opts.TargetDir)
	if err != nil {
		return nil, nil, fmt.Errorf("resolving target dir: %w", err)
	}

	err = fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if opts.SkipPaths[path] {
			return nil
		}

		target := filepath.Join(opts.TargetDir, path)

		absFile, absErr := filepath.Abs(target)
		if absErr != nil {
			return fmt.Errorf("resolving path %s: %w", path, absErr)
		}
		if !strings.HasPrefix(absFile, absTarget+string(filepath.Separator)) && absFile != absTarget {
			return fmt.Errorf("template path %q escapes target directory", path)
		}

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		content, err := fs.ReadFile(src, path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		hash := HashBytes(content)
		hashes[path] = hash

		if _, statErr := os.Lstat(target); statErr == nil {
			if !opts.Overwrite {
				result.Skipped = append(result.Skipped, path)
				return nil
			}
			result.Overwritten = append(result.Overwritten, path)
		} else {
			result.Created = append(result.Created, path)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("creating parent dir for %s: %w", path, err)
		}

		perm := filePerm(path)

		return os.WriteFile(target, content, perm)
	})

	return result, hashes, err
}

// FilePerm returns the appropriate file permission for the given path.
// Shell scripts (.sh suffix) and the extensionless "ralph" script in
// scripts/ get 0755; all others get 0644.
func FilePerm(path string) fs.FileMode {
	if strings.HasSuffix(path, ".sh") {
		return 0755
	}
	if filepath.Base(path) == "ralph" && strings.Contains(path, "scripts/") {
		return 0755
	}
	return 0644
}

// filePerm is an unexported alias for internal use in this package.
func filePerm(path string) fs.FileMode { return FilePerm(path) }

func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", h)
}

func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}
