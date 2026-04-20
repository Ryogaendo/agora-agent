# DESIGN.md — agora-agent Web UI

> agora-agent: Claude Managed Agents を使ったクロスリポジトリ分析 + 共有知識管理ツール

## 1. Visual Theme & Atmosphere

**哲学**: 「知識の広場（Agora）」—— 静かで集中できるが、知識が流れている活気がある空間。
ターミナルの実用性と、ダッシュボードの一覧性を両立する。

**Key Characteristics**:
- **ダーク基調**: 長時間の分析作業に適した目に優しい配色。ターミナルライクだが洗練されている
- **最小限の装飾**: ボーダーとスペーシングで構造を表現。影やグラデーションは控えめ
- **モノスペースの活用**: コード、ログ、エージェント出力はモノスペースで。見出しと操作はサンセリフ
- **紫のアクセント**: Anthropic / Claude のブランドカラーを控えめに使用。操作可能な要素に限定
- **角丸は控えめ**: `border-radius: 6px` を基本。丸すぎず、シャープすぎず
- **情報密度: 中〜高**: エンジニア向けツール。余白は適切だが、スカスカにはしない

## 2. Color Palette & Roles

### Background
- **BG Primary** (`#0f1117`): `--bg-primary`, メイン背景
- **BG Secondary** (`#161822`): `--bg-secondary`, カード・パネル背景
- **BG Tertiary** (`#1e2030`): `--bg-tertiary`, 入力欄・コードブロック背景
- **BG Hover** (`#252840`): `--bg-hover`, ホバー状態

### Text
- **Text Primary** (`#e2e4eb`): `--text-primary`, 本文・見出し
- **Text Secondary** (`#8b8fa3`): `--text-secondary`, 補足・ラベル
- **Text Muted** (`#555970`): `--text-muted`, プレースホルダー・無効状態
- **Text Inverse** (`#0f1117`): `--text-inverse`, アクセントボタン上のテキスト

### Accent
- **Accent** (`#a855f7`): `--accent`, 主要アクション、リンク、選択状態
- **Accent Hover** (`#9333ea`): `--accent-hover`, アクセントのホバー
- **Accent Soft** (`rgba(168, 85, 247, 0.12)`): `--accent-soft`, アクセント背景
- **Accent Border** (`rgba(168, 85, 247, 0.3)`): `--accent-border`, アクセントボーダー

### Semantic
- **Success** (`#34d399`): `--success`, 完了・正常
- **Warning** (`#fbbf24`): `--warning`, 注意・実行中
- **Error** (`#f87171`): `--error`, エラー・失敗
- **Info** (`#60a5fa`): `--info`, 情報・ツール使用中

### Border
- **Border Default** (`#2a2d3e`): `--border`, 通常のボーダー
- **Border Active** (`#3d4158`): `--border-active`, フォーカス・アクティブ

## 3. Typography Rules

| Role | Font | Size | Weight | Line Height | Letter Spacing | CSS Variable | Notes |
|------|------|------|--------|-------------|----------------|-------------|-------|
| Page Title | var(--sans) | 24px | 600 | 1.25 | -0.02em | `--text-page-title` | ページの最上位見出し |
| Section Title | var(--sans) | 18px | 600 | 1.35 | -0.01em | `--text-section-title` | セクション見出し |
| Body | var(--sans) | 14px | 400 | 1.60 | 0 | `--text-body` | 本文 |
| Body Small | var(--sans) | 13px | 400 | 1.50 | 0 | `--text-body-sm` | 補足テキスト |
| Label | var(--sans) | 12px | 500 | 1.30 | 0.03em | `--text-label` | フォームラベル、タブ |
| Code / Log | var(--mono) | 13px | 400 | 1.55 | 0 | `--text-code` | エージェント出力、コードブロック |
| Code Small | var(--mono) | 12px | 400 | 1.50 | 0 | `--text-code-sm` | インラインコード、バッジ |

**Fonts**:
- `--sans`: `'Inter', system-ui, -apple-system, sans-serif`
- `--mono`: `'JetBrains Mono', 'Fira Code', ui-monospace, monospace`

## 4. Component Stylings

### Button — Primary
```
default:    bg: var(--accent)          text: var(--text-inverse)  border: none            radius: 6px  padding: 8px 16px
hover:      bg: var(--accent-hover)    text: var(--text-inverse)  border: none            radius: 6px
active:     bg: #7c22ce               text: var(--text-inverse)  border: none            radius: 6px  scale: 0.98
disabled:   bg: var(--bg-tertiary)     text: var(--text-muted)    border: none            radius: 6px  opacity: 0.6
```

### Button — Secondary
```
default:    bg: transparent            text: var(--text-primary)  border: 1px solid var(--border)          radius: 6px  padding: 8px 16px
hover:      bg: var(--bg-hover)        text: var(--text-primary)  border: 1px solid var(--border-active)   radius: 6px
active:     bg: var(--bg-tertiary)     text: var(--text-primary)  border: 1px solid var(--border-active)   radius: 6px  scale: 0.98
disabled:   bg: transparent            text: var(--text-muted)    border: 1px solid var(--border)          radius: 6px  opacity: 0.5
```

### Button — Ghost
```
default:    bg: transparent            text: var(--text-secondary)  border: none   radius: 6px  padding: 6px 12px
hover:      bg: var(--bg-hover)        text: var(--text-primary)    border: none   radius: 6px
```

### Input / Textarea
```
default:    bg: var(--bg-tertiary)     text: var(--text-primary)    border: 1px solid var(--border)          radius: 6px  padding: 10px 12px
focus:      bg: var(--bg-tertiary)     text: var(--text-primary)    border: 1px solid var(--accent-border)   radius: 6px  outline: none
            box-shadow: 0 0 0 2px var(--accent-soft)
placeholder: color: var(--text-muted)
```

### Card / Panel
```
default:    bg: var(--bg-secondary)    border: 1px solid var(--border)    radius: 8px    padding: 16px
hover:      bg: var(--bg-secondary)    border: 1px solid var(--border-active)
```

### Chat Message — Agent
```
container:  bg: var(--bg-secondary)    border: 1px solid var(--border)    radius: 8px    padding: 14px 16px
text:       font: var(--text-body), color: var(--text-primary)
code:       font: var(--text-code), bg: var(--bg-tertiary), radius: 4px, padding: 2px 6px
```

### Chat Message — User
```
container:  bg: var(--accent-soft)     border: 1px solid var(--accent-border)    radius: 8px    padding: 14px 16px
text:       font: var(--text-body), color: var(--text-primary)
```

### Tool Use Badge
```
container:  bg: var(--bg-tertiary)     border: 1px solid var(--border)    radius: 4px    padding: 4px 8px
icon:       color: var(--info)         size: 14px
text:       font: var(--text-code-sm), color: var(--text-secondary)
```

### Sidebar Navigation
```
item-default:   bg: transparent             text: var(--text-secondary)    padding: 8px 12px   radius: 6px
item-hover:     bg: var(--bg-hover)         text: var(--text-primary)
item-active:    bg: var(--accent-soft)      text: var(--accent)            border-left: 2px solid var(--accent)
```

### Status Indicator
```
running:    color: var(--warning)      animation: pulse 2s infinite
idle:       color: var(--success)
error:      color: var(--error)
```

## 5. Layout Principles

**Spacing Scale** (4px base):
| Token | Value | CSS Variable | Use |
|-------|-------|-------------|-----|
| xs | 4px | `--space-xs` | インラインの隙間 |
| sm | 8px | `--space-sm` | コンポーネント内パディング |
| md | 12px | `--space-md` | フォーム要素間 |
| base | 16px | `--space-base` | カード内パディング、セクション間 |
| lg | 24px | `--space-lg` | セクション間 |
| xl | 32px | `--space-xl` | ページセクション間 |
| 2xl | 48px | `--space-2xl` | ページ上下マージン |

**Layout**:
- **Sidebar + Main**: サイドバー 240px 固定 + メインコンテンツ fluid
- **Container Max**: 960px（メインコンテンツ）
- **Chat Container**: 720px max-width, 中央揃え

## 6. Depth & Elevation

| Level | CSS Variable | Value | Use |
|-------|-------------|-------|-----|
| None | — | none | ほとんどの要素（ボーダーで構造化） |
| Subtle | `--shadow-subtle` | `0 1px 3px rgba(0,0,0,0.2)` | ドロップダウン |
| Medium | `--shadow-md` | `0 4px 12px rgba(0,0,0,0.3)` | モーダル、トースト |
| Glow | `--shadow-glow` | `0 0 12px var(--accent-soft)` | フォーカスリング、強調 |

**哲学**: ダークテーマでは影よりボーダーが構造を作る。影は浮遊要素（ドロップダウン、モーダル）にのみ使用。

## 7. Do's and Don'ts

### Do
- `var(--bg-secondary)` でカード背景を統一する。ハードコードしない
- エージェント出力は必ず `var(--mono)` で表示する
- ツール使用（bash, web_fetch 等）は `Tool Use Badge` コンポーネントで統一表示
- ストリーミング中は `Status Indicator` の `running` アニメーションを表示する
- アクセントカラーは操作可能な要素にのみ使用する

### Don't
- ライトテーマを作らない（v1 ではダークのみ）
- エージェント出力にマークダウンレンダリングを入れすぎない（コードブロック + テキストで十分）
- 角丸を 8px 以上にしない（カード: 8px が最大、それ以外: 6px）
- 紫以外のアクセントカラーを追加しない
- アニメーションを 200ms 以上にしない（`transition: all 150ms ease`）

## 8. Responsive Behavior

| Breakpoint | Width | Layout Change |
|-----------|-------|---------------|
| Desktop | ≥ 1024px | サイドバー + メイン |
| Tablet | 768-1023px | サイドバーをオーバーレイに |
| Mobile | < 768px | シングルカラム、サイドバー非表示 |

- **タッチターゲット**: 最低 44px
- **サイドバー**: tablet 以下でハンバーガーメニュー → オーバーレイ
- **チャット入力**: 常にビューポート下部に固定
- **フォントサイズ**: mobile でも 14px を下回らない

## 9. Agent Prompt Guide

### Quick Color Reference
```
Background:  #0f1117 → #161822 → #1e2030 → #252840  (深 → 浅)
Text:        #e2e4eb → #8b8fa3 → #555970              (明 → 暗)
Accent:      #a855f7 (purple)
Success:     #34d399 (green)
Warning:     #fbbf24 (yellow)
Error:       #f87171 (red)
Info:        #60a5fa (blue)
```

### Example Component Prompts

**Chat Page**:
"Create a chat interface with dark background var(--bg-primary). Left sidebar 240px
with nav items using var(--bg-hover) on hover. Main area max-width 720px centered.
User messages have var(--accent-soft) background with var(--accent-border) border.
Agent messages have var(--bg-secondary) background. Input fixed to bottom with
var(--bg-tertiary) background, var(--border) border, focus ring var(--accent-soft).
All text 14px Inter, agent output in 13px JetBrains Mono."

**Session List**:
"Create a session list as cards in var(--bg-secondary) with var(--border) border,
8px radius, 16px padding. Each card shows title in 14px weight 600, status badge
(running=yellow, idle=green, error=red) in 12px mono, and timestamp in 13px
var(--text-secondary). Cards have hover state var(--border-active)."

**Tool Use Indicator**:
"Inline badge with var(--bg-tertiary) background, var(--border) border, 4px radius.
Left icon 14px in var(--info) blue. Text in 12px JetBrains Mono var(--text-secondary).
Example: [🔧 web_fetch] or [⚙️ bash]. Appears inline within agent message flow."

**Skill Sync Page**:
"Grid of skill cards, 2 columns on desktop, 1 on mobile. Each card var(--bg-secondary)
with skill name in 14px weight 600, description in 13px var(--text-secondary),
sync status badge. Synced skills have var(--success) dot, unsynced have var(--text-muted)
dot. Action button 'Sync All' primary purple at top right."
