package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ryogaendo/agora-agent/internal/agent"
	"github.com/Ryogaendo/agora-agent/internal/config"
	"github.com/spf13/cobra"
)

var (
	updateDomain string
	updateURL    string
	updateSkill  string
)

var updateCmd = &cobra.Command{
	Use:   "update [url]",
	Short: "Fetch a URL and add knowledge to agora/",
	Long: `Read a URL (blog post, documentation, paper) and generate a
structured knowledge document in agora/.

Examples:
  agora-agent update --domain engineering "https://claude.com/blog/..."
  agora-agent update --domain science --skill theoria "https://arxiv.org/..."`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateDomain, "domain", "engineering", "agora/ subdirectory (engineering, science, business)")
	updateCmd.Flags().StringVar(&updateSkill, "skill", "theoria", "Skill to apply (theoria, arkhe, mental-model)")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	url := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := agent.NewClient(agent.ClientConfig{
		AgentID:       cfg.AgentID,
		EnvironmentID: cfg.EnvironmentID,
		GitHubToken:   cfg.GitHubToken,
	})

	// Build prompt based on skill
	prompt := buildUpdatePrompt(url, updateSkill, updateDomain)

	fmt.Fprintf(os.Stderr, "Fetching and analyzing: %s\n", url)
	fmt.Fprintf(os.Stderr, "Skill: %s, Domain: %s\n", updateSkill, updateDomain)

	events, err := client.Run(ctx, agent.RunParams{
		Prompt: prompt,
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

	// Auto-save to agora/
	filename := slugify(url) + ".md"
	outPath := filepath.Join(cfg.AgoraPath, updateDomain, filename)
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, []byte(output.String()), 0644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Saved to %s\n", outPath)

	return nil
}

func buildUpdatePrompt(url, skill, domain string) string {
	date := time.Now().Format("2006-01-02")
	base := fmt.Sprintf(`以下の URL の内容を読み、構造化されたドキュメントを生成してください。

URL: %s
取得日: %s

重要なルール:
- ファイルに書き込まないでください（write ツールを使わない）
- ドキュメントの全文をそのままテキストとして出力してください
- 「生成しました」「以下にまとめます」等の前置きや要約は不要です
- マークダウン形式のドキュメント本文だけを出力してください

`, url, date)

	switch skill {
	case "theoria":
		return base + `以下の構成で出力してください:

# {タイトル}

出典: {URL}
取得日: {日付}
生成スキル: /theoria

---

## 概要
{一言で何か}

## 位置づけ
{隣接概念との比較}

## 原理
{なぜ効くか、第一原理}

## 構造
{レイヤー分解、内部の仕組み}

## 転写メモ
{自分のプロジェクトへの適用可能性}

## 関連
{関連する技術・概念へのリンク}`

	case "arkhe":
		return base + `第一原理まで分解してください。以下の構成で:

# {タイトル}

出典: {URL}
取得日: {日付}
生成スキル: /arkhe

---

## 一言で
## 第一原理
## なぜ存在するか
## 代替手段
## この実装の制約
## 判断基準`

	default:
		return base + `概要、原理、構造、転写メモ（自分のプロジェクトへの適用）を含めてください。`
	}
}

func slugify(url string) string {
	// Extract meaningful part from URL
	parts := strings.Split(url, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		if part != "" && part != "/" {
			// Clean up
			part = strings.ReplaceAll(part, " ", "-")
			part = strings.ToLower(part)
			if len(part) > 60 {
				part = part[:60]
			}
			return part
		}
	}
	return "untitled"
}
