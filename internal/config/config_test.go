package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if !cfg.Doctor.RequireClaudeCLI {
		t.Error("require_claude_cli should be true by default")
	}
	// Codex is opt-in: defaults to false so projects without Codex
	// installed do not see `ralph doctor` exit non-zero.
	if cfg.Doctor.RequireCodexCLI {
		t.Error("require_codex_cli should default to false")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/ralph.toml")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if !cfg.Doctor.RequireClaudeCLI {
		t.Error("require_claude_cli should be true by default")
	}
}

func TestLoad_FullRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ralph.toml")

	content := `[doctor]
require_claude_cli = true
require_go = false
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !cfg.Doctor.RequireClaudeCLI {
		t.Error("require_claude_cli should be true")
	}
	if cfg.Doctor.RequireGo {
		t.Error("require_go should be false")
	}
}

// TestLoad_TemplateRalphToml verifies the shipped templates/base/ralph.toml
// parses cleanly under the current Config schema.
func TestLoad_TemplateRalphToml(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file location")
	}
	path := filepath.Join(filepath.Dir(thisFile), "..", "..", "templates", "base", "ralph.toml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("templates/base/ralph.toml not found: %v", err)
	}
	if _, err := Load(path); err != nil {
		t.Fatalf("Load(templates/base/ralph.toml): %v", err)
	}
}

// TestLoad_RequireCodexCLI verifies the require_codex_cli field round-trips.
func TestLoad_RequireCodexCLI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ralph.toml")

	content := `[doctor]
require_claude_cli = false
require_codex_cli = true
require_go = false
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Doctor.RequireCodexCLI {
		t.Error("require_codex_cli = false after loading explicit true")
	}
	if cfg.Doctor.RequireClaudeCLI {
		t.Error("require_claude_cli = true after loading explicit false")
	}
}
