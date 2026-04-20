package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Ryogaendo/agora-agent/internal/agent"
	"github.com/Ryogaendo/agora-agent/internal/config"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure agora-agent (create Agent and Environment on Anthropic)",
	RunE:  runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Check API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Println("ANTHROPIC_API_KEY is not set.")
		fmt.Println("Set it in your shell: export ANTHROPIC_API_KEY=sk-ant-...")
		return fmt.Errorf("missing ANTHROPIC_API_KEY")
	}
	fmt.Println("✓ ANTHROPIC_API_KEY is set")

	// GitHub token
	if cfg.GitHubToken == "" {
		fmt.Print("GitHub token (for repo access, leave empty to skip): ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)
		if token != "" {
			cfg.GitHubToken = token
		}
	} else {
		fmt.Println("✓ GitHub token is configured")
	}

	// Create Agent
	client := agent.NewClient(agent.ClientConfig{
		GitHubToken: cfg.GitHubToken,
	})

	systemPrompt := cfg.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = `You are a cross-repository analysis agent.

You have access to multiple repositories and a shared knowledge store (agora/).
When analyzing code, consider cross-project implications and shared patterns.`
	}

	fmt.Println("\nCreating Agent on Anthropic...")
	agentID, err := client.EnsureAgent(ctx, "agora-agent", systemPrompt)
	if err != nil {
		return err
	}
	cfg.AgentID = agentID
	fmt.Printf("✓ Agent created: %s\n", agentID)

	fmt.Println("Creating Environment...")
	envID, err := client.EnsureEnvironment(ctx)
	if err != nil {
		return err
	}
	cfg.EnvironmentID = envID
	fmt.Printf("✓ Environment created: %s\n", envID)

	// Save config
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("\n✓ Config saved to %s\n", config.DefaultConfigPath())

	fmt.Println("\nReady! Try:")
	fmt.Println("  agora-agent analyze --repos <repo> \"Summarize the monorepo structure\"")

	return nil
}
