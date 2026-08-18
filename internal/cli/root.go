package cli

import (
	"github.com/spf13/cobra"
)

// Version, GitCommit, and BuildDate are set via ldflags at build time.
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "ralph",
		Short:         "Claude Code harness scaffold and pipeline CLI",
		Long:          `ralph scaffolds Claude Code projects with best-practice configurations, hooks, skills, and pipeline settings.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newInitCmd(),
		newUpgradeCmd(),
		newDoctorCmd(),
		newPackCmd(),
		newVersionCmd(),
		newInsightsCmd(),
	)

	return root
}
