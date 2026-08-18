package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"

	"github.com/thomas0124/ralph/internal/config"
	"github.com/thomas0124/ralph/internal/scaffold"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check environment and project setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(".")
		},
	}
}

type checkResult struct {
	Name   string `json:"name"`
	Status string `json:"status"` // pass, info, warn, fail
	Detail string `json:"detail,omitempty"`
}

func runDoctor(targetDir string) error {
	cfg, cfgErr := config.Load(filepath.Join(targetDir, "ralph.toml"))
	var results []checkResult

	if cfgErr != nil && !errors.Is(cfgErr, fs.ErrNotExist) {
		results = append(results, checkResult{
			Name:   "ralph.toml",
			Status: "warn",
			Detail: fmt.Sprintf("parse error: %v — using defaults", cfgErr),
		})
	}

	results = append(results, checkClaudeCLI(cfg))
	results = append(results, checkCodexCLI(cfg))
	results = append(results, checkCodexEffectiveConfig(targetDir))
	results = append(results, checkGo(cfg))
	results = append(results, checkHooks(targetDir))
	results = append(results, checkManifestVersion(targetDir))
	results = append(results, checkInstalledPacks(targetDir)...)

	fmt.Println("ralph doctor")
	fmt.Println()

	allPass := true
	for _, r := range results {
		icon := "✓"
		switch r.Status {
		case "info":
			icon = "ℹ"
		case "warn":
			icon = "⚠"
		case "fail":
			icon = "✗"
			allPass = false
		}
		fmt.Printf("  %s %s: %s", icon, r.Name, r.Status)
		if r.Detail != "" {
			fmt.Printf(" — %s", r.Detail)
		}
		fmt.Println()
	}

	fmt.Println()
	if allPass {
		fmt.Println("All checks passed.")
		return nil
	}
	fmt.Println("Some checks failed. Fix the issues above.")
	return fmt.Errorf("doctor: %d check(s) failed", countFailed(results))
}

func countFailed(results []checkResult) int {
	n := 0
	for _, r := range results {
		if r.Status == "fail" {
			n++
		}
	}
	return n
}

func probeBinary(bin string) (string, error) {
	if _, lookErr := exec.LookPath(bin); lookErr != nil {
		return "", fmt.Errorf("%s not found in PATH", bin)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, runErr := exec.CommandContext(ctx, bin, "--version").CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("%s --version timed out after 5s", bin)
	}
	if runErr != nil {
		return "", fmt.Errorf("%s --version failed: %w", bin, runErr)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed, nil
		}
	}
	return "", fmt.Errorf("%s --version produced no output", bin)
}

func checkClaudeCLI(cfg config.Config) checkResult {
	r := checkResult{Name: "Claude Code CLI"}
	version, err := probeBinary("claude")
	if err != nil {
		if cfg.Doctor.RequireClaudeCLI {
			r.Status = "fail"
			r.Detail = fmt.Sprintf("claude unusable: %v", err)
		} else {
			r.Status = "warn"
			r.Detail = fmt.Sprintf("claude unusable (not required): %v", err)
		}
		return r
	}
	r.Status = "pass"
	r.Detail = version
	return r
}

func checkCodexCLI(cfg config.Config) checkResult {
	r := checkResult{Name: "Codex"}
	version, err := probeBinary("codex")
	if err != nil {
		if cfg.Doctor.RequireCodexCLI {
			r.Status = "fail"
			r.Detail = fmt.Sprintf("codex unusable: %v", err)
		} else {
			r.Status = "warn"
			r.Detail = fmt.Sprintf("codex unusable (not required): %v", err)
		}
		return r
	}
	r.Status = "pass"
	r.Detail = version
	return r
}

// checkCodexEffectiveConfig confirms that .codex/config.toml is present and
// carries the bits Codex actually loads from a project-level config:
// [features] hooks = true plus at least one [hooks.<event>] entry.
// We cannot probe Codex's trust state from Go, so the result stays a warning
// when the file is structurally fine — the user has to confirm trust via
// `codex trust .`.
func checkCodexEffectiveConfig(targetDir string) checkResult {
	r := checkResult{Name: "Codex effective config"}
	cfgPath := filepath.Join(targetDir, ".codex", "config.toml")
	data, err := os.ReadFile(cfgPath)
	if errors.Is(err, fs.ErrNotExist) {
		r.Status = "warn"
		r.Detail = ".codex/config.toml not found"
		return r
	}
	if err != nil {
		r.Status = "warn"
		r.Detail = fmt.Sprintf("could not read .codex/config.toml: %v", err)
		return r
	}

	var raw map[string]any
	if err := toml.Unmarshal(data, &raw); err != nil {
		r.Status = "fail"
		r.Detail = fmt.Sprintf("invalid .codex/config.toml: %v", err)
		return r
	}

	hooksFeatureEnabled := false
	if features, ok := raw["features"].(map[string]any); ok {
		if v, ok := features["hooks"].(bool); ok {
			hooksFeatureEnabled = v
		}
	}

	hookEntries := 0
	hasInlineHooks := false
	if hooks, ok := raw["hooks"].(map[string]any); ok {
		hasInlineHooks = true
		for _, eventHooks := range hooks {
			switch v := eventHooks.(type) {
			case []any:
				hookEntries += len(v)
			case map[string]any:
				if len(v) > 0 {
					hookEntries++
				}
			}
		}
	}

	hooksJSONPath := filepath.Join(targetDir, ".codex", "hooks.json")
	if hasInlineHooks {
		if _, err := os.Stat(hooksJSONPath); err == nil {
			r.Status = "fail"
			r.Detail = "both .codex/config.toml [hooks] and .codex/hooks.json exist; remove hooks.json"
			return r
		} else if !errors.Is(err, fs.ErrNotExist) {
			r.Status = "warn"
			r.Detail = fmt.Sprintf("could not inspect .codex/hooks.json: %v", err)
			return r
		}
	}

	switch {
	case !hooksFeatureEnabled:
		r.Status = "warn"
		r.Detail = "[features] hooks = true is not set; project hooks will be ignored"
	case hookEntries == 0:
		r.Status = "warn"
		r.Detail = "no [hooks.*] entries — run `codex trust .` once configured"
	default:
		r.Status = "pass"
		r.Detail = fmt.Sprintf("hooks=true, %d hook entry(ies). Confirm `codex trust .` ran for this project", hookEntries)
	}
	return r
}

func checkGo(cfg config.Config) checkResult {
	r := checkResult{Name: "Go"}
	_, err := exec.LookPath("go")
	if err != nil {
		if cfg.Doctor.RequireGo {
			r.Status = "fail"
			r.Detail = "go not found in PATH"
		} else {
			r.Status = "info"
			r.Detail = "not installed (not required)"
		}
	} else {
		out, runErr := exec.Command("go", "version").Output()
		if runErr != nil {
			r.Status = "pass"
		} else {
			r.Status = "pass"
			r.Detail = strings.TrimSpace(string(out))
		}
	}
	return r
}

func checkHooks(targetDir string) checkResult {
	r := checkResult{Name: "Hooks integrity"}
	settingsPath := filepath.Join(targetDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		r.Status = "warn"
		r.Detail = "settings.json not found"
		return r
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		r.Status = "fail"
		r.Detail = "invalid settings.json"
		return r
	}

	hooks, ok := settings["hooks"]
	if !ok {
		r.Status = "warn"
		r.Detail = "no hooks configured"
		return r
	}

	hooksMap, ok := hooks.(map[string]any)
	if !ok {
		r.Status = "pass"
		return r
	}

	missing := 0
	for _, eventHooks := range hooksMap {
		eventList, ok := eventHooks.([]any)
		if !ok {
			continue
		}
		for _, eh := range eventList {
			ehMap, ok := eh.(map[string]any)
			if !ok {
				continue
			}
			hooksList, ok := ehMap["hooks"].([]any)
			if !ok {
				continue
			}
			for _, h := range hooksList {
				hMap, ok := h.(map[string]any)
				if !ok {
					continue
				}
				cmd, ok := hMap["command"].(string)
				if !ok {
					continue
				}
				fields := strings.Fields(cmd)
				if len(fields) == 0 {
					continue
				}
				exe := fields[0]
				if _, err := os.Lstat(filepath.Join(targetDir, exe)); errors.Is(err, fs.ErrNotExist) {
					missing++
				}
			}
		}
	}

	if missing > 0 {
		r.Status = "fail"
		r.Detail = fmt.Sprintf("%d hook script(s) missing", missing)
	} else {
		r.Status = "pass"
	}
	return r
}

func checkManifestVersion(targetDir string) checkResult {
	r := checkResult{Name: "Manifest version"}
	manifestPath := filepath.Join(targetDir, ".ralph", "manifest.toml")
	m, err := scaffold.ReadManifest(manifestPath)
	if err != nil {
		r.Status = "warn"
		r.Detail = "no manifest found — run 'ralph init'"
		return r
	}

	if m.Meta.Version == Version {
		r.Status = "pass"
		r.Detail = m.Meta.Version
	} else {
		r.Status = "warn"
		r.Detail = fmt.Sprintf("manifest %s ≠ CLI %s — run 'ralph upgrade'", m.Meta.Version, Version)
	}
	return r
}

func checkInstalledPacks(targetDir string) []checkResult {
	manifestPath := filepath.Join(targetDir, ".ralph", "manifest.toml")
	m, err := scaffold.ReadManifest(manifestPath)
	if err != nil {
		return nil
	}

	if len(m.Meta.Packs) == 0 {
		return []checkResult{{Name: "Language packs", Status: "pass", Detail: "none installed"}}
	}

	var results []checkResult
	for _, p := range m.Meta.Packs {
		r := checkResult{Name: fmt.Sprintf("Pack: %s", p)}
		verifyPath := filepath.Join(targetDir, "packs", "languages", p, "verify.sh")
		relVerify := filepath.Join("packs", "languages", p, "verify.sh")
		if _, err := os.Stat(verifyPath); errors.Is(err, fs.ErrNotExist) {
			r.Status = "warn"
			r.Detail = fmt.Sprintf("%s not found on disk", relVerify)
		} else {
			r.Status = "pass"
		}
		results = append(results, r)
	}
	return results
}
