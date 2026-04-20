package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agora-agent",
	Short: "Cross-repository analysis agent powered by Claude Managed Agents",
	Long: `agora-agent is a CLI that uses Claude Managed Agents to analyze
codebases, generate knowledge, and maintain the shared knowledge store (agora/).

It knows about your project structure, can mount multiple repositories,
and automatically stores results in agora/.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
