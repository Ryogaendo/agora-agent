package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/Ryogaendo/agora-agent/internal/config"
	"github.com/Ryogaendo/agora-agent/internal/routine"
	"github.com/spf13/cobra"
)

// --- root ---

var routineCmd = &cobra.Command{
	Use:   "routine",
	Short: "Manage Claude Code routines for automated agora maintenance",
	Long: `Register, trigger, and manage Claude Code routines.

Routines run on Anthropic cloud infrastructure on a schedule, via API,
or in response to GitHub events. Use this command group to wire them
into the agora-agent workflow.

Setup:
  1. Create a routine at claude.ai/code/routines (use "agora-agent routine templates" for prompts)
  2. Add an API trigger and copy the trigger ID + token
  3. Register it:  agora-agent routine add freshness-check --trigger-id trig_01... --token sk-ant-oat01-...
  4. Fire it:      agora-agent routine fire freshness-check`,
}

func init() {
	rootCmd.AddCommand(routineCmd)
}

// --- add ---

var (
	addTriggerID   string
	addToken       string
	addDescription string
)

var routineAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Register a routine's API trigger",
	Long: `Store a routine's trigger ID and bearer token so you can fire it by name.

Example:
  agora-agent routine add freshness-check \
    --trigger-id trig_01ABCDEFGHJKLMNOPQRSTUVW \
    --token sk-ant-oat01-xxxxx \
    --desc "Weekly agora freshness check"`,
	Args: cobra.ExactArgs(1),
	RunE: runRoutineAdd,
}

func init() {
	routineAddCmd.Flags().StringVar(&addTriggerID, "trigger-id", "", "Routine trigger ID (required)")
	routineAddCmd.Flags().StringVar(&addToken, "token", "", "Bearer token for the trigger (required)")
	routineAddCmd.Flags().StringVar(&addDescription, "desc", "", "Short description")
	routineAddCmd.MarkFlagRequired("trigger-id")
	routineAddCmd.MarkFlagRequired("token")
	routineCmd.AddCommand(routineAddCmd)
}

func runRoutineAdd(_ *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Routines == nil {
		cfg.Routines = make(map[string]config.Routine)
	}

	if _, exists := cfg.Routines[name]; exists {
		fmt.Fprintf(os.Stderr, "Overwriting existing routine %q\n", name)
	}

	cfg.Routines[name] = config.Routine{
		TriggerID:   addTriggerID,
		Token:       addToken,
		Description: addDescription,
	}

	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("✓ Routine %q registered\n", name)
	return nil
}

// --- fire ---

var fireText string

var routineFireCmd = &cobra.Command{
	Use:   "fire <name> [--text '...']",
	Short: "Trigger a registered routine",
	Long: `POST to the routine's /fire endpoint and return the session URL.

Examples:
  agora-agent routine fire freshness-check
  agora-agent routine fire alert-triage --text "Sentry alert SEN-4521 in prod"`,
	Args: cobra.ExactArgs(1),
	RunE: runRoutineFire,
}

func init() {
	routineFireCmd.Flags().StringVar(&fireText, "text", "", "Run-specific context passed to the routine")
	routineCmd.AddCommand(routineFireCmd)
}

func runRoutineFire(_ *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	r, ok := cfg.Routines[name]
	if !ok {
		return fmt.Errorf("routine %q not found (run: agora-agent routine list)", name)
	}

	fmt.Fprintf(os.Stderr, "Firing routine %q ...\n", name)

	resp, err := routine.Fire(r.TriggerID, r.Token, fireText)
	if err != nil {
		return err
	}

	fmt.Printf("Session: %s\n", resp.SessionID)
	fmt.Printf("URL:     %s\n", resp.SessionURL)
	return nil
}

// --- list ---

var routineListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered routines",
	RunE:  runRoutineList,
}

func init() {
	routineCmd.AddCommand(routineListCmd)
}

func runRoutineList(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Routines) == 0 {
		fmt.Println("No routines registered. Use: agora-agent routine add <name> ...")
		return nil
	}

	names := make([]string, 0, len(cfg.Routines))
	for n := range cfg.Routines {
		names = append(names, n)
	}
	sort.Strings(names)

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTRIGGER ID\tDESCRIPTION")
	for _, n := range names {
		r := cfg.Routines[n]
		// Mask the trigger ID for readability
		tid := r.TriggerID
		if len(tid) > 16 {
			tid = tid[:12] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", n, tid, r.Description)
	}
	w.Flush()
	return nil
}

// --- remove ---

var routineRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a registered routine",
	Args:  cobra.ExactArgs(1),
	RunE:  runRoutineRemove,
}

func init() {
	routineCmd.AddCommand(routineRemoveCmd)
}

func runRoutineRemove(_ *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, ok := cfg.Routines[name]; !ok {
		return fmt.Errorf("routine %q not found", name)
	}

	delete(cfg.Routines, name)
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("✓ Routine %q removed\n", name)
	return nil
}

// --- templates ---

var routineTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Show pre-built routine prompts for agora maintenance",
	Long: `Print ready-to-use prompts for common agora maintenance routines.

Copy a prompt into claude.ai/code/routines when creating a new routine,
then register the API trigger with: agora-agent routine add <name> ...`,
	RunE: runRoutineTemplates,
}

func init() {
	routineCmd.AddCommand(routineTemplatesCmd)
}

type tmpl struct {
	Name        string
	Schedule    string
	Description string
	Prompt      string
}

var templates = []tmpl{
	{
		Name:     "freshness-check",
		Schedule: "Weekly (Sunday 10:00)",
		Description: "Scan agora articles and flag stale entries",
		Prompt: strings.TrimSpace(`
あなたは agora/ (共有知識ストア) の鮮度を管理するエージェントです。

## タスク

1. agora/ 配下のすべての .md ファイルを読む
2. 各ファイルの「取得日」フィールドを確認する
3. 以下の基準で分類する:
   - 🟢 Fresh: 3ヶ月以内
   - 🟡 Aging: 3〜6ヶ月
   - 🔴 Stale: 6ヶ月以上
4. 🔴 Stale の記事について、内容が現在も有効かを技術的に判断する
5. レポートを生成する（Markdown テーブル形式）

## 出力

以下の形式でレポートを出力:
- ファイルパス | 取得日 | ステータス | 推奨アクション
- Stale 記事のうち更新が必要なものはPRを作成

## 注意
- 削除は提案のみ。実際の削除はしない
- PDF やノートブックはスキップ
`),
	},
	{
		Name:     "docs-drift",
		Schedule: "Weekly (Monday 09:00)",
		Description: "Detect agora articles that reference changed APIs",
		Prompt: strings.TrimSpace(`
あなたは agora/ の記事とコードベースの整合性を検証するエージェントです。

## タスク

1. 過去1週間にマージされた PR を各リポジトリで確認する
2. 変更されたファイル・API・設定を特定する
3. agora/ の記事で、それらに言及しているものを検索する
4. 言及内容が変更後のコードと矛盾していないか検証する

## 出力

- 矛盾が検出された場合: 記事の更新PRを作成
- 矛盾がない場合: 簡潔なサマリーのみ
`),
	},
	{
		Name:     "weekly-digest",
		Schedule: "Weekly (Friday 17:00)",
		Description: "Summarize the week's changes and suggest new agora articles",
		Prompt: strings.TrimSpace(`
あなたは週次の技術ダイジェストを生成するエージェントです。

## タスク

1. 今週マージされた PR を全リポジトリで収集する
2. 以下の観点で分類・要約する:
   - 新機能
   - バグ修正
   - リファクタリング / インフラ変更
   - 依存関係の更新
3. 複数プロジェクトに共通する変更パターンを検出する
4. agora/ に蓄積すべき知見があれば記事のドラフトを作成する

## 出力形式

# 週次ダイジェスト (YYYY-MM-DD)

## ハイライト
- ...

## プロジェクト別サマリー
{リポジトリごとのセクション}

## 横断的な知見
{複数プロジェクトに共通するパターンや学び}

## 提案: agora 新記事
{ドラフトがあれば PR を作成}
`),
	},
	{
		Name:     "pr-review",
		Schedule: "GitHub trigger: pull_request.opened",
		Description: "Bespoke code review with team checklist",
		Prompt: strings.TrimSpace(`
あなたはチーム固有のコードレビューチェックリストを適用するエージェントです。

## レビュー観点

### セキュリティ
- [ ] SQLインジェクション / XSS / CSRF の防御
- [ ] 認証・認可の適切な実装
- [ ] シークレットのハードコード禁止
- [ ] 入力バリデーション

### パフォーマンス
- [ ] N+1 クエリ
- [ ] 不要な再レンダリング (React)
- [ ] インデックスの考慮 (DB変更時)

### 設計
- [ ] 責務の分離（Single Responsibility）
- [ ] agora/ の既存パターンとの整合性
- [ ] エラーハンドリングの一貫性

### テスト
- [ ] 新機能のテストカバレッジ
- [ ] エッジケースの考慮
- [ ] テストの可読性

## 出力
- PR にインラインコメントを追加
- サマリーコメントで全体評価を記載
- Critical / Warning / Info のラベルを使用
`),
	},
}

func runRoutineTemplates(_ *cobra.Command, _ []string) error {
	for i, t := range templates {
		if i > 0 {
			fmt.Println("\n" + strings.Repeat("─", 60) + "\n")
		}
		fmt.Printf("## %s\n", t.Name)
		fmt.Printf("Schedule:    %s\n", t.Schedule)
		fmt.Printf("Description: %s\n\n", t.Description)
		fmt.Println("```")
		fmt.Println(t.Prompt)
		fmt.Println("```")
	}

	fmt.Printf("\n%s\n\n", strings.Repeat("─", 60))
	fmt.Println("Usage:")
	fmt.Println("  1. Copy a prompt above into claude.ai/code/routines")
	fmt.Println("  2. Add an API trigger and copy the trigger ID + token")
	fmt.Println("  3. Register: agora-agent routine add <name> --trigger-id <id> --token <token>")
	fmt.Println("  4. Fire:     agora-agent routine fire <name>")
	return nil
}
