package main

import (
	"fmt"
	"os"

	ralph "github.com/thomas0124/ralph"
	"github.com/thomas0124/ralph/internal/cli"
	"github.com/thomas0124/ralph/internal/scaffold"
)

func main() {
	cli.Version = Version
	cli.GitCommit = GitCommit
	cli.BuildDate = BuildDate
	scaffold.EmbeddedFS = ralph.TemplatesFS

	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
