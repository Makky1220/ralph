package scaffold

import (
	"os"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

type Manifest struct {
	Meta  ManifestMeta            `toml:"meta"`
	Files map[string]ManifestFile `toml:"files"`
}

type ManifestMeta struct {
	Version string   `toml:"version"`
	Created string   `toml:"created"`
	Updated string   `toml:"updated"`
	Packs   []string `toml:"packs,omitempty"`
}

type ManifestFile struct {
	Hash    string `toml:"hash"`
	Managed bool   `toml:"managed"`

	State          string `toml:"state,omitempty"`
	TemplateHash   string `toml:"template_hash,omitempty"`
	DiskHash       string `toml:"disk_hash,omitempty"`
	BaselineStatus string `toml:"baseline_status,omitempty"`
	BaselinePath   string `toml:"baseline_path,omitempty"`
}

const (
	FileStateManaged   = "managed"
	FileStateUnmanaged = "unmanaged"
	FileStatePartial   = "partial"

	BaselineStatusMissing   = "missing"
	BaselineStatusAvailable = "available"
)

func NewManifest(version string) *Manifest {
	now := time.Now().UTC().Format(time.RFC3339)
	return &Manifest{
		Meta: ManifestMeta{
			Version: version,
			Created: now,
			Updated: now,
		},
		Files: make(map[string]ManifestFile),
	}
}

func ReadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Files == nil {
		m.Files = make(map[string]ManifestFile)
	}
	return &m, nil
}

func (m *Manifest) Write(path string) error {
	m.Meta.Updated = time.Now().UTC().Format(time.RFC3339)
	data, err := toml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (m *Manifest) SetFile(relPath, hash string) {
	m.Files[relPath] = ManifestFile{
		Hash:           hash,
		Managed:        true,
		State:          FileStateManaged,
		TemplateHash:   hash,
		BaselineStatus: BaselineStatusMissing,
	}
}

func (m *Manifest) SetFileWithBaseline(relPath, hash, baselinePath string) {
	m.Files[relPath] = ManifestFile{
		Hash:           hash,
		Managed:        true,
		State:          FileStateManaged,
		TemplateHash:   hash,
		BaselineStatus: BaselineStatusAvailable,
		BaselinePath:   baselinePath,
	}
}

func (m *Manifest) SetFileResolvedWithBaseline(relPath, templateHash, diskHash, state, baselinePath string) {
	m.Files[relPath] = ManifestFile{
		Hash:           templateHash,
		Managed:        true,
		State:          state,
		TemplateHash:   templateHash,
		DiskHash:       diskHash,
		BaselineStatus: BaselineStatusAvailable,
		BaselinePath:   baselinePath,
	}
}

func (m *Manifest) SetFileUnmanaged(relPath, hash string) {
	m.Files[relPath] = ManifestFile{
		Hash:           hash,
		Managed:        false,
		State:          FileStateUnmanaged,
		DiskHash:       hash,
		BaselineStatus: BaselineStatusMissing,
	}
}

func (f ManifestFile) IsBaselineAvailable() bool {
	return f.BaselineStatus == BaselineStatusAvailable && f.BaselinePath != ""
}

func (f ManifestFile) WithTemplateHash(hash string) ManifestFile {
	f.Hash = hash
	f.Managed = true
	if f.State == "" {
		f.State = FileStateManaged
	}
	if f.TemplateHash == "" {
		f.TemplateHash = hash
	}
	if f.BaselineStatus == "" {
		f.BaselineStatus = BaselineStatusMissing
	}
	return f
}
