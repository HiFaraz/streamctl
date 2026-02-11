# streamctl

**Persistent memory for AI coding agents.**

When you're using Claude Code, Cursor, or other AI assistants for complex multi-session work, context disappears between sessions. What were you working on? What decisions did you make? Where did you leave off?

streamctl solves this with **workstreams** - persistent units of work that survive across sessions, track progress, and preserve decisions.

## Why?

AI coding assistants are great for single-session tasks. But real projects span days or weeks:

- **Monday**: "Implement authentication" - you make progress, discuss trade-offs, decide on JWT
- **Tuesday**: New session. The AI has no memory. You re-explain everything. Again.
- **Wednesday**: Different approach emerges. Why did you reject it on Monday? Nobody knows.

With streamctl:

```
# Monday
workstream_create(project="myapp", name="auth", objective="JWT authentication")
workstream_update(name="auth", log_entry="Chose JWT over sessions - need stateless for Lambda")
workstream_update(name="auth", task_add="Token validation middleware")
workstream_update(name="auth", task_status={"position": 0, "status": "done"})

# Tuesday - new session picks up instantly
workstream_get(project="myapp", name="auth")
→ Shows objective, tasks, decisions, exactly where you left off
```

## Features

- **Task tracking** with status (pending/in_progress/done/skipped) and markdown notes
- **Decision log** - record why you chose X over Y, never re-litigate
- **Dependencies** - mark workstreams as blocked by others
- **needs_help flag** - signal when you're stuck and need human attention
- **Live web dashboard** - monitor parallel agents in real-time
- **Keyboard-native UI** - navigate with `.`/`,`, `/` to search, `?` for help
- **Export to markdown** - sync to git via pre-commit hooks
- **Works with any MCP client** - Claude Code, Claude Desktop, etc.

## Quick Start

In Claude Code:

```
Clone https://github.com/HiFaraz/streamctl to ~/streamctl, build with `go build -o streamctl ./cmd/streamctl`, then run `claude mcp add streamctl --scope user -- ~/streamctl/streamctl serve`
```

Then tell Claude to use workstreams in your `~/.claude/CLAUDE.md`:

```markdown
## Workstream Management

At session start, check `workstream_list(project="<repo>")` and resume any in_progress work.
During work, log decisions and progress. At session end, update state and note what's next.
```

## Web Dashboard

Monitor workstreams in your browser:

```
web_serve(project="myproject")
→ http://localhost:54321
```

Live-updating feed of activity across all workstreams. Keyboard-native navigation. Ideal for watching parallel agents work.

**Keyboard shortcuts**: `.`/`,` navigate, `Enter` opens, `/` searches, `Backspace` goes back, `?` shows help.

## Use Cases

### Solo Development

Track complex features across sessions:

```
workstream_create(project="myapp", name="refactor-db", objective="Migrate from MySQL to Postgres")
workstream_update(name="refactor-db", log_entry="Schema converted. Testing migration script tomorrow.")
```

Next session reads the log and continues.

### Parallel Agents

Run multiple Claude Code instances on different workstreams:

```
# Agent 1                          # Agent 2
workstream_claim("api")            workstream_claim("frontend")
# works on API...                  # works on frontend...
workstream_update(log_entry="...")  workstream_update(log_entry="...")
```

Watch both in the web dashboard. Flag `needs_help=true` when stuck.

### Team Coordination

Break work into independent streams, track dependencies:

```
workstream_create(name="auth", objective="Authentication system")
workstream_create(name="api", objective="REST endpoints")
workstream_update(name="api", add_blocker="myapp/auth")  # api waits for auth
```

Dashboard shows what's blocked and why.

## CLI Commands

```bash
streamctl serve              # Start MCP server (for Claude Code)
streamctl web                # Open web dashboard
streamctl tui                # Terminal UI
streamctl export PROJECT     # Export to markdown (for git)
streamctl list               # JSON dump
```

## MCP Tools

| Tool | Description |
|------|-------------|
| `workstream_list` | List workstreams, filter by project/state/owner |
| `workstream_get` | Full workstream details as markdown |
| `workstream_create` | Create new workstream |
| `workstream_update` | Update state, log, tasks, dependencies, needs_help |
| `workstream_claim` | Set ownership |
| `workstream_release` | Clear ownership |
| `web_serve` | Start web dashboard, returns URL |

### workstream_update Parameters

| Parameter | Description |
|-----------|-------------|
| `state` | pending, in_progress, blocked, done |
| `log_entry` | Append timestamped note (supports markdown) |
| `task_add` | Add task |
| `task_status` | `{"position": 0, "status": "done"}` |
| `task_notes` | `{"position": 0, "notes": "markdown here"}` |
| `add_blocker` | `"project/name"` - mark as blocked by |
| `needs_help` | `true` - flag for human attention |

## Export to Git

Keep workstreams in version control:

```bash
streamctl export myproject --dir ./workstreams/
```

Add to pre-commit hook:

```bash
#!/bin/bash
streamctl export myproject --dir ./workstreams/
git add workstreams/*.md
```

Exported files are marked as generated - edit via streamctl, not directly.

## Installation

Requires Go 1.21+:

```bash
git clone https://github.com/HiFaraz/streamctl ~/streamctl
cd ~/streamctl
go build -o streamctl ./cmd/streamctl
./streamctl init
```

Add to Claude Code:

```bash
claude mcp add streamctl --scope user -- ~/streamctl/streamctl serve
```

## How It Works

streamctl is an [MCP server](https://modelcontextprotocol.io/) that exposes workstream tools to AI assistants. Data is stored in SQLite at `~/.streamctl/workstreams.db`.

The mental model:
- **Workstream** = a unit of work spanning multiple sessions (like an epic)
- **Tasks** = checklist items within a workstream
- **Log** = timestamped decisions and progress notes
- **State** = pending → in_progress → done (or blocked)

## Contributing

Feedback welcome. Open an issue or append to `~/streamctl/FEEDBACK.md` while using it.

## License

MIT
