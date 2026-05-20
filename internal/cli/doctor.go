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

	"github.com/spf13/cobra"

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
	Name   string
	Status string
	Detail string
}

func runDoctor(targetDir string) error {
	var results []checkResult

	results = append(results, checkClaudeCLI())
	results = append(results, checkHooks(targetDir))
	results = append(results, checkManifestVersion(targetDir))
	results = append(results, checkInstalledPacks(targetDir)...)

	fmt.Println("ralph doctor")
	fmt.Println()

	allPass := true
	for _, r := range results {
		icon := "✓"
		switch r.Status {
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

func checkClaudeCLI() checkResult {
	r := checkResult{Name: "Claude Code CLI"}
	version, err := probeBinary("claude")
	if err != nil {
		r.Status = "fail"
		r.Detail = fmt.Sprintf("claude unusable: %v", err)
		return r
	}
	r.Status = "pass"
	r.Detail = version
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
				if _, err := os.Stat(filepath.Join(targetDir, cmd)); errors.Is(err, fs.ErrNotExist) {
					missing++
				}
			}
		}
	}

	if missing > 0 {
		r.Status = "fail"
		r.Detail = fmt.Sprintf("%d hook script(s) missing on disk", missing)
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
		if _, err := os.Stat(verifyPath); errors.Is(err, fs.ErrNotExist) {
			r.Status = "warn"
			r.Detail = "verify.sh not found on disk"
		} else {
			r.Status = "pass"
		}
		results = append(results, r)
	}
	return results
}
