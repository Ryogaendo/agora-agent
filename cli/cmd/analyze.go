package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ryogaendo/agora-agent/internal/agent"
	"github.com/Ryogaendo/agora-agent/internal/config"
	"github.com/spf13/cobra"
)

var (
	analyzeRepos  []string
	analyzeOutput string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [prompt]",
	Short: "Analyze one or more repositories with a prompt",
	Long: `Mount repositories into a Managed Agents session and run analysis.
Results are streamed to stdout and optionally saved to agora/.

Examples:
  agora-agent analyze --repos repo-a,repo-b "Compare authentication flows"
  agora-agent analyze --repos repo-a --output engineering/auth-analysis.md "Review the auth module"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAnalyze,
}

func init() {
	analyzeCmd.Flags().StringSliceVar(&analyzeRepos, "repos", nil, "Repositories to mount (comma-separated, e.g. repo-a,repo-b)")
	analyzeCmd.Flags().StringVar(&analyzeOutput, "output", "", "Save output to agora/<path> (e.g. engineering/analysis.md)")
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	prompt := strings.Join(args, " ")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client := agent.NewClient(agent.ClientConfig{
		AgentID:       cfg.AgentID,
		EnvironmentID: cfg.EnvironmentID,
		GitHubToken:   cfg.GitHubToken,
	})

	// Build repo mounts
	var repos []agent.RepoMount
	for _, name := range analyzeRepos {
		url, ok := cfg.Repos[name]
		if !ok {
			return fmt.Errorf("unknown repo %q (known: %v)", name, repoNames(cfg.Repos))
		}
		repos = append(repos, agent.RepoMount{
			URL:       url,
			MountPath: fmt.Sprintf("/workspace/%s", name),
		})
	}

	fmt.Fprintf(os.Stderr, "Starting session with %d repo(s)...\n", len(repos))

	events, err := client.Run(ctx, agent.RunParams{
		Prompt: prompt,
		Repos:  repos,
	})
	if err != nil {
		return err
	}

	var output strings.Builder
	for ev := range events {
		switch ev.Type {
		case "message":
			fmt.Print(ev.Text)
			output.WriteString(ev.Text)
		case "tool_use":
			fmt.Fprintf(os.Stderr, "\n[Using tool: %s]\n", ev.Text)
		case "done":
			fmt.Fprintf(os.Stderr, "\n\nDone.\n")
		case "error":
			return fmt.Errorf("agent error: %s", ev.Text)
		}
	}

	// Save to agora/ if --output specified
	if analyzeOutput != "" {
		outPath := filepath.Join(cfg.AgoraPath, analyzeOutput)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("creating output dir: %w", err)
		}
		if err := os.WriteFile(outPath, []byte(output.String()), 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Saved to %s\n", outPath)
	}

	return nil
}

func repoNames(repos map[string]string) []string {
	names := make([]string, 0, len(repos))
	for k := range repos {
		names = append(names, k)
	}
	return names
}
