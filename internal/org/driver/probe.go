package driver

import (
	"context"
	"fmt"
)

// ProbeModel launches a minimal, side-effect-free invocation of drv with
// model to confirm the CLI accepts that model id. A non-nil error (with
// stderr detail folded in by the Runner, see ExecRunner.Run) means the
// model id was rejected or the CLI itself failed to run.
//
// codex `--model` support on `codex exec` is best-effort (known upstream
// gap: some codex builds ignore or reject --model on the exec subcommand),
// so a ProbeModel failure for drv=="codex" is advisory, not fatal -- the
// caller (checkOrgModelProbes in internal/cli/doctor.go, wired into
// `ralph doctor --probe-models`) decides how to surface that severity.
//
// opencode models use the provider/model form (e.g.
// "anthropic/claude-sonnet-4-5") and are passed verbatim to `opencode run
// --model`; a failure is reported as a plain warn (no advisory special-case).
func ProbeModel(ctx context.Context, r Runner, drv, model string) error {
	switch drv {
	case "claude":
		if _, err := r.Run(ctx, "claude", "--model", model, "-p", "ping", "--output-format", "text"); err != nil {
			return fmt.Errorf("probe claude model %q: %w", model, err)
		}
		return nil
	case "codex":
		if _, err := r.Run(ctx, "codex", "exec", "--model", model, "--skip-git-repo-check", "ping"); err != nil {
			return fmt.Errorf("probe codex model %q: %w", model, err)
		}
		return nil
	case "opencode":
		if _, err := r.Run(ctx, "opencode", "run", "--model", model, "ping"); err != nil {
			return fmt.Errorf("probe opencode model %q: %w", model, err)
		}
		return nil
	default:
		return fmt.Errorf("probe: unknown driver %q (want claude|codex|opencode)", drv)
	}
}
