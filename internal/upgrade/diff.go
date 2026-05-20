package upgrade

import (
	"io/fs"
	"path/filepath"

	"github.com/thomas0124/ralph/internal/scaffold"
)

type DiffOptions struct {
	CheckRemovals bool
	SkipPaths     map[string]bool
}

type FileAction int

const (
	ActionAutoUpdate FileAction = iota
	ActionConflict
	ActionAdd
	ActionRemove
	ActionSkip
)

type FileDiff struct {
	Path       string
	Action     FileAction
	OldHash    string
	DiskHash   string
	NewHash    string
	NewContent []byte
}

func ComputeDiffsWithManifest(manifest *scaffold.Manifest, targetDir string, newFS fs.FS, checkRemovals bool) ([]FileDiff, error) {
	return ComputeDiffsWithManifestOptions(manifest, targetDir, newFS, DiffOptions{
		CheckRemovals: checkRemovals,
	})
}

func ComputeDiffsWithManifestOptions(manifest *scaffold.Manifest, targetDir string, newFS fs.FS, opts DiffOptions) ([]FileDiff, error) {
	var diffs []FileDiff

	newFiles := make(map[string]bool)
	err := fs.WalkDir(newFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if opts.SkipPaths[path] {
			return nil
		}
		newFiles[path] = true

		content, err := fs.ReadFile(newFS, path)
		if err != nil {
			return err
		}
		appendFileDiff(&diffs, manifest, targetDir, path, content)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if opts.CheckRemovals {
		for path, mf := range manifest.Files {
			if newFiles[path] || opts.SkipPaths[path] {
				continue
			}
			if !mf.Managed {
				diffs = append(diffs, FileDiff{
					Path:    path,
					Action:  ActionSkip,
					OldHash: mf.Hash,
				})
				continue
			}
			diffs = append(diffs, FileDiff{
				Path:    path,
				Action:  ActionRemove,
				OldHash: mf.Hash,
			})
		}
	}

	return diffs, nil
}

func ComputeFileDiff(manifest *scaffold.Manifest, targetDir, path string, content []byte) FileDiff {
	var diffs []FileDiff
	appendFileDiff(&diffs, manifest, targetDir, path, content)
	return diffs[0]
}

func appendFileDiff(diffs *[]FileDiff, manifest *scaffold.Manifest, targetDir, path string, content []byte) {
	newHash := scaffold.HashBytes(content)

	mf, inManifest := manifest.Files[path]

	diskPath := filepath.Join(targetDir, path)
	diskHash, diskErr := scaffold.HashFile(diskPath)

	if inManifest && !mf.Managed {
		*diffs = append(*diffs, FileDiff{
			Path:       path,
			Action:     ActionSkip,
			OldHash:    mf.Hash,
			DiskHash:   diskHash,
			NewHash:    newHash,
			NewContent: content,
		})
		return
	}

	if !inManifest {
		if diskErr == nil && diskHash != newHash {
			*diffs = append(*diffs, FileDiff{
				Path:       path,
				Action:     ActionConflict,
				OldHash:    newHash,
				DiskHash:   diskHash,
				NewHash:    newHash,
				NewContent: content,
			})
			return
		}
		*diffs = append(*diffs, FileDiff{
			Path:       path,
			Action:     ActionAdd,
			NewHash:    newHash,
			NewContent: content,
		})
		return
	}

	if diskErr != nil {
		*diffs = append(*diffs, FileDiff{
			Path:       path,
			Action:     ActionAdd,
			NewHash:    newHash,
			NewContent: content,
		})
		return
	}

	if mf.Hash == "" {
		if diskHash == newHash {
			*diffs = append(*diffs, FileDiff{
				Path:       path,
				Action:     ActionSkip,
				OldHash:    mf.Hash,
				DiskHash:   diskHash,
				NewHash:    newHash,
				NewContent: content,
			})
			return
		}
		*diffs = append(*diffs, FileDiff{
			Path:       path,
			Action:     ActionConflict,
			OldHash:    newHash,
			DiskHash:   diskHash,
			NewHash:    newHash,
			NewContent: content,
		})
		return
	}

	recordedTemplateHash := mf.Hash
	if mf.TemplateHash != "" {
		recordedTemplateHash = mf.TemplateHash
	}
	recordedDiskHash := mf.Hash
	if mf.DiskHash != "" {
		recordedDiskHash = mf.DiskHash
	}

	if newHash == recordedTemplateHash {
		if diskHash == recordedDiskHash {
			*diffs = append(*diffs, FileDiff{
				Path:       path,
				Action:     ActionSkip,
				OldHash:    recordedTemplateHash,
				DiskHash:   diskHash,
				NewHash:    newHash,
				NewContent: content,
			})
			return
		}
		*diffs = append(*diffs, FileDiff{
			Path:       path,
			Action:     ActionConflict,
			OldHash:    recordedTemplateHash,
			DiskHash:   diskHash,
			NewHash:    newHash,
			NewContent: content,
		})
		return
	}

	if diskHash == recordedTemplateHash {
		*diffs = append(*diffs, FileDiff{
			Path:       path,
			Action:     ActionAutoUpdate,
			OldHash:    recordedTemplateHash,
			DiskHash:   diskHash,
			NewHash:    newHash,
			NewContent: content,
		})
	} else {
		*diffs = append(*diffs, FileDiff{
			Path:       path,
			Action:     ActionConflict,
			OldHash:    recordedTemplateHash,
			DiskHash:   diskHash,
			NewHash:    newHash,
			NewContent: content,
		})
	}
}
