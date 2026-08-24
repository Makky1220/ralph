package cli

import (
	"strings"
	"testing"

	"github.com/thomas0124/ralph/internal/config"
)

// TestCheckOpencodeCLI_Present pins the happy path: an opencode binary whose
// --version exits 0 reports pass with the version line as detail.
func TestCheckOpencodeCLI_Present(t *testing.T) {
	dir := t.TempDir()
	writeStubBin(t, dir, "opencode", "")
	t.Setenv("PATH", dir)

	r := checkOpencodeCLI(config.Default())
	if r.Status != "pass" {
		t.Fatalf("checkOpencodeCLI status = %q, want pass (detail=%q)", r.Status, r.Detail)
	}
	if r.Name != "OpenCode" {
		t.Errorf("checkOpencodeCLI name = %q, want OpenCode", r.Name)
	}
	if r.Detail != "pong" {
		t.Errorf("checkOpencodeCLI detail = %q, want the stub's version output", r.Detail)
	}
}

// TestCheckOpencodeCLI_Absent_NotRequired pins the default (opt-in) posture:
// opencode missing from PATH is only a warn when require_opencode_cli is
// false — doctor's exit code stays clean.
func TestCheckOpencodeCLI_Absent_NotRequired(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	cfg := config.Default()
	cfg.Doctor.RequireOpencodeCLI = false

	r := checkOpencodeCLI(cfg)
	if r.Status != "warn" {
		t.Fatalf("checkOpencodeCLI status = %q, want warn (detail=%q)", r.Status, r.Detail)
	}
	if !strings.Contains(r.Detail, "not required") {
		t.Errorf("detail %q should mention the not-required fallback", r.Detail)
	}
}

// TestCheckOpencodeCLI_Absent_Required pins the operator opt-in: with
// [doctor].require_opencode_cli = true a missing binary is a hard fail.
func TestCheckOpencodeCLI_Absent_Required(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	cfg := config.Default()
	cfg.Doctor.RequireOpencodeCLI = true

	r := checkOpencodeCLI(cfg)
	if r.Status != "fail" {
		t.Fatalf("checkOpencodeCLI status = %q, want fail (detail=%q)", r.Status, r.Detail)
	}
}

// TestCheckOpencodeCLI_BrokenShim_RequiredFails mirrors the claude/codex
// broken-shim contract for opencode: LookPath success alone must not report
// pass; probeBinary surfaces the --version failure and require_opencode_cli
// escalates it to fail.
func TestCheckOpencodeCLI_BrokenShim_RequiredFails(t *testing.T) {
	dir := t.TempDir()
	writeFailingStubBin(t, dir, "opencode", "shim broken")
	t.Setenv("PATH", dir)

	cfg := config.Default()
	cfg.Doctor.RequireOpencodeCLI = true

	r := checkOpencodeCLI(cfg)
	if r.Status != "fail" {
		t.Fatalf("checkOpencodeCLI status = %q, want fail for a broken shim under require_opencode_cli (detail=%q)", r.Status, r.Detail)
	}
}
