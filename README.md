# agora-agent

Cross-repository analysis agent CLI powered by [Anthropic Managed Agents](https://docs.anthropic.com/). Mount multiple repositories, run Claude against them with a shared knowledge store (`agora/`), and register Claude Code routines for scheduled maintenance.

> `agora/` is the shared knowledge store — a directory of Markdown notes produced by the agent and consumed by future runs. Point `agora_path` in the config at any directory you want to use for this purpose.

## Features

- **`analyze`** — mount one or more GitHub repos into a Managed Agents session and run a prompt; optionally save the result into `agora/`.
- **`update`** — fetch a URL (blog post, docs, paper) and generate a structured Markdown document in `agora/` via a skill (`theoria`, `arkhe`, …).
- **`skill sync`** — upload local skills (Claude Code skill directories) to Anthropic so the agent can call them.
- **`routine`** — register, fire, and manage [Claude Code routines](https://claude.com/claude-code/routines) by name. Ships with template prompts for freshness checks, docs drift detection, weekly digests, and PR review.

## Requirements

- Go 1.22+
- An Anthropic API key (`ANTHROPIC_API_KEY`)
- A GitHub token with repo read access (optional, required only when mounting private repos)

## Install

```bash
git clone https://github.com/Ryogaendo/agora-agent.git
cd agora-agent/cli
go build -o agora-agent .
# move the binary somewhere on your PATH, e.g.:
mv agora-agent ~/bin/
```

## Setup

```bash
export ANTHROPIC_API_KEY=sk-ant-...
agora-agent setup
```

This creates a Managed Agent and Environment on Anthropic and writes the IDs to `~/.agora-agent.json`. You can edit that file to:

- set `agora_path` (default: `~/projects/agora`)
- register repos under `repos` (name → GitHub URL)
- customize `system_prompt` for the agent

## Usage

```bash
# Analyze one or more repos
agora-agent analyze --repos web,api "Compare the authentication flows"
agora-agent analyze --repos web --output engineering/auth.md "Review the auth module"

# Fetch a URL and store structured notes in agora/
agora-agent update --domain engineering "https://example.com/post"
agora-agent update --domain science --skill theoria "https://arxiv.org/abs/..."

# Skills
agora-agent skill sync                    # sync all local skills
agora-agent skill sync theoria techne     # sync specific skills
agora-agent skill list

# Routines
agora-agent routine templates              # show ready-to-use prompts
agora-agent routine add <name> --trigger-id trig_01... --token sk-ant-oat01-...
agora-agent routine fire <name> --text "optional per-run context"
agora-agent routine list
```

## Repository layout

```
cli/       Go CLI (cobra)
  cmd/         sub-commands (analyze, update, setup, skill, routine)
  internal/    agent client, config, routine client
web/       optional web UI (TanStack Start + Nitro, WIP)
```

## Configuration

`~/.agora-agent.json`:

```json
{
  "agent_id": "agt_...",
  "environment_id": "env_...",
  "github_token": "ghp_... (optional)",
  "agora_path": "/Users/you/projects/agora",
  "system_prompt": "You are a cross-repository analysis agent...",
  "repos": {
    "web": "https://github.com/you/web",
    "api": "https://github.com/you/api"
  },
  "skills": {
    "theoria": "skl_..."
  },
  "routines": {
    "freshness-check": { "trigger_id": "trig_...", "token": "sk-ant-oat01-..." }
  }
}
```

The file is created by `setup` with mode `0600`.

## License

MIT
