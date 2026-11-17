# kie-cli

Command-line client and MCP server for the [KIE.ai](https://kie.ai) generation API.
Create image and video jobs (Nano Banana Pro, Sora 2, Kling and other market models), wait for results and check credits from a terminal, a script or an AI agent.

- **One binary, two interfaces:** `kie-cli` for shells and scripts, `kie-mcp` for MCP-compatible agents.
- **Built for automation:** JSON output, field selection, dry runs, no interactive prompts in agent mode.
- **Predictable exit codes**, so pipelines can react to auth, rate-limit and API errors.

## Install

Requires Go 1.26 or later.

```bash
go install github.com/BKproduct/kie-cli/cmd/kie-cli@latest
go install github.com/BKproduct/kie-cli/cmd/kie-mcp@latest   # optional MCP server
kie-cli --version
```

## Authenticate

Get an API key at <https://kie.ai/api-key>, then either export it or store it:

```bash
export KIE_API_KEY="your-key"        # checked first; KIE_BEARER_AUTH also works
kie-cli auth set-token "your-key"    # or keep it in the local config file
kie-cli doctor                       # checks config, credentials and connectivity
```

The config file lives at `~/.config/kie-cli/config.toml`.

## Quick start

```bash
# Remaining credits
kie-cli chat

# Create an image job, then wait for the result
kie-cli jobs create-task --model nano-banana-pro \
  --input '{"prompt":"studio photo of a ceramic vase","aspect_ratio":"1:1"}' --json
kie-cli jobs wait --task-id TASK_ID --select data.state,data.resultJson

# Temporary download link for a generated file
kie-cli common --help
```

## Commands

| Command | What it does |
|---|---|
| `jobs create-task` | Submit an async generation task (`--model`, `--input` JSON or `--stdin`) |
| `jobs record-info` | Get task status and result |
| `jobs wait` | Poll a task until success, failure or timeout |
| `chat` | Show remaining account credits |
| `common` | Get a temporary download URL for a generated file |
| `doctor` | Check configuration, credentials and connectivity |
| `which "<capability>"` | Find the command for a task described in plain words |
| `profile` | Save and reuse named sets of flags |
| `agent-context` | Describe the CLI as JSON for agents |

Run `kie-cli <command> --help` for every flag.

## Output and scripting

```bash
kie-cli chat --json                       # JSON
kie-cli chat --json --select id,status    # only the fields you need
kie-cli jobs create-task ... --dry-run    # show the request without sending it
kie-cli chat --agent                      # JSON, compact, no prompts, no colour
```

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Use with AI agents

**Claude Code:**

```bash
claude mcp add kie -- kie-mcp
```

**Claude Desktop** (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "kie": {
      "command": "kie-mcp",
      "env": { "KIE_BEARER_AUTH": "<your-key>" }
    }
  }
}
```

`SKILL.md` in this repository is a ready-made agent skill: link the repo into your agent's skills folder to use it.

## Development

```bash
make build   # builds bin/kie-cli
make test
```

## Licence

Apache 2.0, see [LICENSE](LICENSE). Third-party notices: [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES).
