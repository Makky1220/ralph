package scaffold

import (
	"fmt"
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

	// v3 metadata. Additive; absent on v1/v2 manifests and on manifests
	// written through the legacy (non-opt-in) constructors/setters below.
	Layout string `toml:"layout,omitempty"`
}

type ManifestFile struct {
	Hash    string `toml:"hash"`
	Managed bool   `toml:"managed"`

	State          string `toml:"state,omitempty"`
	TemplateHash   string `toml:"template_hash,omitempty"`
	DiskHash       string `toml:"disk_hash,omitempty"`
	BaselineStatus string `toml:"baseline_status,omitempty"`
	BaselinePath   string `toml:"baseline_path,omitempty"`

	// v3 metadata (ownership). Additive; absent (legacy/unset) unless
	// written through the opt-in v3 setters below.
	Owner             string `toml:"owner,omitempty"`
	ForkedFromVersion string `toml:"forked_from_version,omitempty"`
}

const (
	FileStateManaged   = "managed"
	FileStateUnmanaged = "unmanaged"
	FileStatePartial   = "partial"

	BaselineStatusMissing   = "missing"
	BaselineStatusAvailable = "available"

	// LayoutV2 identifies the overlay-scaffold layout (manifest v3, ownership
	// attributes). Set only through the opt-in SetLayoutV2 API.
	LayoutV2 = "v2"

	// Ownership attributes for manifest v3 entries.
	OwnerCore  = "core"
	OwnerFork  = "fork"
	OwnerSeed  = "seed"
	OwnerBlock = "block"
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

// SetLayoutV2 marks the manifest as belonging to the overlay-scaffold (v2)
// layout. Opt-in only: existing constructors/setters never set this.
func (m *Manifest) SetLayoutV2() {
	m.Meta.Layout = LayoutV2
}

// SetFileOwned records a manifest v3 entry with an explicit ownership
// attribute (core/seed/block). Fork entries are rejected: SetFileFork is
// the single way to record a fork.
func (m *Manifest) SetFileOwned(relPath, owner, templateHash, diskHash string) error {
	if err := validateManagedOwner(owner); err != nil {
		return err
	}
	m.Files[relPath] = ManifestFile{
		Hash:           templateHash,
		Managed:        true,
		State:          FileStateManaged,
		TemplateHash:   templateHash,
		DiskHash:       diskHash,
		BaselineStatus: BaselineStatusMissing,
		Owner:          owner,
	}
	return nil
}

// SetFileFork records a manifest v3 entry for a file the user has ejected
// from core ownership into a fork. Managed is false: upgrade must not
// replace fork content.
func (m *Manifest) SetFileFork(relPath, diskHash, forkedFromVersion string) {
	m.Files[relPath] = ManifestFile{
		Hash:              diskHash,
		Managed:           false,
		State:             FileStateUnmanaged,
		DiskHash:          diskHash,
		BaselineStatus:    BaselineStatusMissing,
		Owner:             OwnerFork,
		ForkedFromVersion: forkedFromVersion,
	}
}

// SetOwner mutates an existing manifest entry's Owner attribute in place.
// Only core/seed/block are accepted here — a fork must be recorded via
// SetFileFork.
func (m *Manifest) SetOwner(relPath, owner string) error {
	if err := validateManagedOwner(owner); err != nil {
		return err
	}
	entry, ok := m.Files[relPath]
	if !ok {
		return fmt.Errorf("no manifest entry for %q", relPath)
	}
	entry.Owner = owner
	m.Files[relPath] = entry
	return nil
}

func validateManagedOwner(owner string) error {
	switch owner {
	case OwnerCore, OwnerSeed, OwnerBlock:
		return nil
	case OwnerFork:
		return fmt.Errorf("owner %q must be recorded via SetFileFork", owner)
	default:
		return fmt.Errorf("unknown owner %q", owner)
	}
}

// IsLegacyOwner reports whether the entry predates manifest v3 ownership
// attribution (owner unset).
func (f ManifestFile) IsLegacyOwner() bool {
	return f.Owner == ""
}
