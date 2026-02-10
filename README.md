# streamctl

Coordinate parallel Claude Code agents working on the same codebase.

## The Problem

When running multiple Claude Code agents in parallel (e.g., one working on auth, another on the API, another on tests), they have no way to:
- Know what work is available or claimed
- Avoid stepping on each other's toes
- Track progress across sessions
- Hand off work between sessions

## The Solution

streamctl provides a shared database of **workstreams** - units of work that agents can claim, update, and complete. Think of it like a task board that Claude Code agents can read and write to.

## Example: Parallel Feature Development

You're building a new feature that needs:
1. Database schema changes
2. API endpoints
3. Frontend components
4. Tests

Instead of doing these sequentially, create workstreams:

```
workstream_create(project="myapp", name="db-schema", objective="Add user preferences table")
workstream_create(project="myapp", name="api-endpoints", objective="CRUD endpoints for preferences")
workstream_create(project="myapp", name="frontend", objective="Preferences settings page")
workstream_create(project="myapp", name="tests", objective="Integration tests for preferences")
```

Now spin up 4 Claude Code agents. Each one:
1. Calls `workstream_list` to see available work
2. Calls `workstream_claim` on an unclaimed workstream
3. Works on it, logging progress with `workstream_update`
4. Marks it done and releases it

No conflicts. Clear ownership. Progress persists across sessions.

## Example: Resuming Work

You worked on a feature yesterday but didn't finish. Today:

```
workstream_list(project="myapp", state="in_progress")
```

Shows your in-progress work with the log of what you did yesterday. Pick up where you left off.

## Example: Blocked Work

Your API work is blocked waiting for the database schema:

```
workstream_update(project="myapp", name="api-endpoints", state="blocked", log_entry="Waiting for db-schema to complete")
```

Another agent (or you tomorrow) can see this is blocked and why.

---

## Setup for Claude Code

```bash
# Clone and build (one-time)
git clone https://github.com/HiFaraz/streamctl ~/streamctl
cd ~/streamctl && go build -o streamctl ./cmd/streamctl

# Initialize in your project
mkdir -p .streamctl
~/streamctl/streamctl init
```

Add to `~/.claude/settings.json`:
```json
{
  "mcpServers": {
    "streamctl": {
      "command": "$HOME/streamctl/streamctl",
      "args": ["serve"]
    }
  }
}
```

Restart Claude Code. Add `.streamctl/` to your `.gitignore`.

---

## MCP Tools

| Tool | Parameters | Description |
|------|------------|-------------|
| `workstream_list` | `project?`, `state?`, `owner?` | List workstreams, optionally filtered |
| `workstream_get` | `project`, `name` | Get full workstream details |
| `workstream_create` | `project`, `name`, `objective` | Create a new workstream |
| `workstream_update` | `project`, `name`, + `state?`, `log_entry?`, `plan_index?` | Update fields |
| `workstream_claim` | `project`, `name`, `owner` | Set ownership (use your session ID) |
| `workstream_release` | `project`, `name` | Clear ownership |

## Workstream Fields

- **project**: Repository or project name
- **name**: Descriptive identifier (e.g., "auth-refactor", "fix-bug-123")
- **state**: `pending` → `in_progress` → `done` (or `blocked`)
- **owner**: Who's working on it (prevents conflicts)
- **objective**: One-sentence goal
- **plan**: Checklist of steps (toggle with `plan_index`)
- **log**: Timestamped progress entries

## CLI

```bash
streamctl tui      # Visual dashboard
streamctl list     # JSON dump of all workstreams
streamctl help     # Help
```
