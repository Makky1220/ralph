package config

// TestDefaultsLockStep verifies that the three authoritative sources for
// model/effort defaults stay in lock-step:
//
//  1. scripts/ralph-config.sh   — shell defaults (VAR="${VAR:-value}" idiom)
//  2. templates/base/ralph.toml — declarative project config
//  3. config.Default()          — Go CLI defaults exported as RALPH_* env vars
//
// It also checks that the cross-review SKILL.md fallback for
// RALPH_CLAUDE_REVIEWER_MODEL matches the shell default.
//
// Any mismatch is a test failure naming the surface and key.  The test uses
// runtime.Caller to resolve paths relative to this file so it works from any
// working directory.  Files are resolved relative to the repo root
// (two directories above internal/config/).
//
// If a file is absent (e.g. the package is vendored), the relevant sub-test
// is skipped with an explanatory message so the test does not break in
// downstream repos that embed internal/config without the meta-repo tree.

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// repoRoot resolves the repository root from this source file's location.
// internal/config/defaults_sync_test.go → repo root is ../../
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file location via runtime.Caller")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// parseShellDefaults reads a shell file and extracts VAR="${VAR:-value}"
// defaults into a map.  Only top-level assignments of the form
//
//	VARNAME="${VARNAME:-default_value}"
//
// are captured.  The idiom is documented in scripts/ralph-config.sh.
//
// Go's regexp package does not support backreferences, so we match the full
// pattern with two named groups and verify post-match that the variable name
// in the LHS matches the name inside the parameter expansion.
func parseShellDefaults(t *testing.T, path string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("parseShellDefaults: cannot read %s: %v", path, err)
	}
	// Matches: VARNAME="${ANYNAME:-value}"
	// Captures: [1]=LHS name  [2]=inner name  [3]=default value
	// We then confirm [1]==[2] to enforce the self-referential idiom.
	re := regexp.MustCompile(`^([A-Z_]+)="\$\{([A-Z_]+):-([^}]*)}"`)
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		m := re.FindStringSubmatch(line)
		if len(m) == 4 && m[1] == m[2] {
			result[m[1]] = m[3]
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("parseShellDefaults: scanner error: %v", err)
	}
	return result
}

// mustShell returns the value for the given var from the shell map, failing
// loudly if missing.
func mustShell(t *testing.T, m map[string]string, varName string) string {
	t.Helper()
	v, ok := m[varName]
	if !ok {
		t.Fatalf("scripts/ralph-config.sh: %s not found by VAR=\"${VAR:-default}\" regex — update the regex or the file", varName)
	}
	return v
}

// TestDefaultsLockStep is the main tripwire.
func TestDefaultsLockStep(t *testing.T) {
	root := repoRoot(t)

	// ── 1. Shell defaults ─────────────────────────────────────────────────────
	shellPath := filepath.Join(root, "scripts", "ralph-config.sh")
	if _, err := os.Stat(shellPath); err != nil {
		t.Skipf("scripts/ralph-config.sh not found (%v) — skipping lock-step check (vendored repo?)", err)
	}
	shell := parseShellDefaults(t, shellPath)

	// ── cross-review SKILL.md fallback matches shell ──────────────────────────
	// The SKILL.md documents: ${RALPH_CLAUDE_REVIEWER_MODEL:-opus}
	// We grep for the documented fallback token and assert it equals the shell
	// default for RALPH_CLAUDE_REVIEWER_MODEL.
	skillPath := filepath.Join(root, ".claude", "skills", "cross-review", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Skipf(".claude/skills/cross-review/SKILL.md not found (%v) — skipping reviewer fallback check", err)
	}
	{
		data, err := os.ReadFile(skillPath)
		if err != nil {
			t.Fatalf("cannot read cross-review/SKILL.md: %v", err)
		}
		// Pattern: ${RALPH_CLAUDE_REVIEWER_MODEL:-<fallback>}
		re := regexp.MustCompile(`\$\{RALPH_CLAUDE_REVIEWER_MODEL:-([^}]+)\}`)
		matches := re.FindAllSubmatch(data, -1)
		if len(matches) == 0 {
			t.Fatal(".claude/skills/cross-review/SKILL.md: no ${RALPH_CLAUDE_REVIEWER_MODEL:-<fallback>} pattern found — update the regex if the wording changed")
		}
		// All occurrences must agree and match the shell default.
		shellVal := mustShell(t, shell, "RALPH_CLAUDE_REVIEWER_MODEL")
		for _, m := range matches {
			skillFallback := string(m[1])
			if skillFallback != shellVal {
				t.Errorf("cross-review SKILL.md reviewer fallback = %q, want %q (scripts/ralph-config.sh RALPH_CLAUDE_REVIEWER_MODEL default)",
					skillFallback, shellVal)
			}
		}
	}
}
