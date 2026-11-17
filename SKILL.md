---
name: kie
description: "KIE.ai jobs API from terminal — createTask, poll, wait, credits. Trigger: `use kie`, `run kie-cli`, `nano-banana-pro on KIE`, `KIE sora`, `kie create task`, `check kie credits`."
author: "Boris Korol"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  requires:
    bins:
      - kie-cli
---

# Kie CLI

## Prerequisites: Install the CLI

This skill drives the `kie-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install with Go 1.26+:
   ```bash
   go install github.com/BKproduct/kie-cli/cmd/kie-cli@latest
   ```
2. Verify: `kie-cli --version`
3. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

If `--version` reports "command not found" after install, the install step did not put the binary on `$PATH`. Do not proceed with skill commands until verification succeeds.

Unified job API for KIE.ai market models (image/video generation).
Bearer token from https://kie.ai/api-key — env KIE_API_KEY.

## When to Use This CLI

Use for **KIE market** image/video generation (NanoBanana Pro, Sora2, Kling, etc.) when `KIE_API_KEY` is set. Replaces one-off generation scripts with agent-native JSON.

## Unique Capabilities

- **`jobs create-task`** — Submit `model` + `input` JSON (same shape as `POST /api/v1/jobs/createTask`).
- **`jobs wait`** — Poll `recordInfo` until `success` / failure / timeout (replaces manual sleep loops).
- **`chat`** — Remaining credits (`8901` style balance).
- **`common`** — Resolve temp download URLs for generated assets.

```bash
export KIE_API_KEY="your-key"   # from https://kie.ai/api-key
kie-cli doctor
kie-cli chat --agent
kie-cli jobs create-task --model nano-banana-pro \
  --input '{"prompt":"…","aspect_ratio":"9:16","output_format":"png","resolution":"2K","image_urls":["https://…"]}' --agent
kie-cli jobs wait --task-id TASK_ID --agent --select data.state,data.resultJson
```

Models (examples): `nano-banana-pro`, `sora-2-text-to-video`, `kling/v2-1-master-image-to-video` — see [KIE docs](https://docs.kie.ai/market/quickstart).

## Command Reference

**jobs** — Async generation tasks (preferred)

- `kie-cli jobs create-task` — Create task (`--model`, `--input` JSON or stdin)
- `kie-cli jobs record-info` — Poll status (`--task-id`)
- `kie-cli jobs wait` — Block until done (`--task-id`, `--max-wait`, `--interval`)

**chat** — Account credits

- `kie-cli chat` — GET `/api/v1/chat/credit`

**common** — File helpers

- `kie-cli common` — POST `/api/v1/common/download-url`


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
kie-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Auth Setup

Run `kie-cli auth setup` for the URL and steps to obtain a token (add `--launch` to open the URL). Then store it:

```bash
kie-cli auth set-token YOUR_TOKEN_HERE
```

Or export **`KIE_API_KEY`** (checked first) or `KIE_BEARER_AUTH`.

Run `kie-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  kie-cli chat --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Non-interactive** — never prompts, every input is a flag
- **Explicit retries** — use `--idempotent` only when an already-existing create should count as success

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
kie-cli feedback "the --since flag is inclusive but docs say exclusive"
kie-cli feedback --stdin < notes.txt
kie-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.kie-cli/feedback.jsonl`. They are never POSTed unless `KIE_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `KIE_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
kie-cli profile save briefing --json
kie-cli --profile briefing chat
kie-cli profile list --json
kie-cli profile show briefing
kie-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `kie-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

Install the MCP binary from this CLI's published public-library entry or pre-built release, then register it:

```bash
claude mcp add kie-mcp -- kie-mcp
```

Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which kie-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   kie-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `kie-cli <command> --help`.
