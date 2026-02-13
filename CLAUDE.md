# streamctl

MCP server + TUI for managing workstreams across projects.

## Build Commands

```bash
make build                    # Build binary
make test                     # Run tests
make hooks                    # Install git hooks (rebuilds binary on commit)
./streamctl init              # Initialize database
./streamctl serve             # Start MCP server
./streamctl tui               # Launch TUI dashboard
./streamctl web               # Start web UI (auto-detects project)
./streamctl list              # List workstreams (JSON)
./streamctl export PROJECT/NAME          # Export single workstream to stdout
./streamctl export PROJECT --dir ./dir/  # Export all to directory
./streamctl log PROJECT "message"        # Log to active workstream
```

## Architecture

```
~/.streamctl/
└── workstreams.db            # SQLite database

streamctl/                    # This repo
├── cmd/streamctl/            # CLI entry point
├── internal/
│   ├── store/                # SQLite storage
│   ├── mcp/                  # MCP server implementation
│   ├── tui/                  # Terminal UI (bubbletea)
│   └── web/                  # Web UI (HTML server)
└── pkg/
    └── workstream/           # Core types and rendering
```

## Database Schema

```sql
workstreams (id, project, name, state, owner, objective, needs_help, last_update, created_at)
plan_items (id, workstream_id, position, text, complete, status, notes)
log_entries (id, workstream_id, timestamp, content)
workstream_dependencies (blocker_id, blocked_id, created_at)
milestones (id, project, name, description, created_at)
milestone_requirements (milestone_id, workstream_id)
```

## MCP Tools

| Tool | Description |
|------|-------------|
| `workstream_list` | List all workstreams, optionally filter by project/state/owner |
| `workstream_get` | Get full workstream content (rendered as markdown) |
| `workstream_create` | Create new workstream |
| `workstream_update` | Update state, log, tasks, dependencies (see below) |
| `workstream_claim` | Set ownership of a workstream |
| `workstream_release` | Release ownership |
| `web_serve` | Start web UI server, returns URL with floating port |
| `milestone_create` | Create a cross-workstream gate/checkpoint |
| `milestone_get` | Get milestone with computed status and requirements |
| `milestone_list` | List milestones with computed status |
| `milestone_update` | Add/remove requirements, update description |

### workstream_update Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `project` | string | Project name (required) |
| `name` | string | Workstream name (required) |
| `new_name` | string | Rename workstream to this name |
| `state` | string | New state: pending, in_progress, blocked, done |
| `log_entry` | string | Append log entry |
| `task_add` | string | Add new task with this text |
| `task_remove` | number | Remove task at position (0-indexed) |
| `task_status` | object | Set task status: `{"position": 0, "status": "done"}` |
| `task_notes` | object | Set task notes (markdown): `{"position": 0, "notes": "..."}` |
| `add_blocker` | string | Add dependency: `project/name` blocks this workstream |
| `remove_blocker` | string | Remove dependency |
| `needs_help` | boolean | Flag workstream as needing help/at-risk |

**Task statuses:** `pending`, `in_progress`, `done`, `skipped`
**Task notes:** Supports markdown (code blocks, lists, links, headers)
**Log entries:** Support markdown for rich context, code snippets, decision rationale

### milestone_update Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `project` | string | Project name (required) |
| `name` | string | Milestone name (required) |
| `description` | string | Update milestone description |
| `add_requirement` | string | Add workstream requirement: `project/name` |
| `remove_requirement` | string | Remove workstream requirement: `project/name` |

**Milestone status** (computed automatically):
- `pending` - no required workstreams are done
- `in_progress` - some (but not all) are done
- `done` - ALL required workstreams have state="done"

## TUI Features

- List all projects and workstreams
- Filter by project, state, owner
- View workstream details
- Keyboard navigation (j/k, Enter, Esc, q)

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `STREAMCTL_DB` | `~/.streamctl/workstreams.db` | Database path |

## Claude Code MCP Configuration

Add to `~/.claude/settings.json`:
```json
{
  "mcpServers": {
    "streamctl": {
      "command": "/path/to/streamctl",
      "args": ["serve"]
    }
  }
}
```

## Web UI

When the user asks to see workstreams in the browser, use `web_serve`:

```
web_serve(project="myproject")
```

This starts an HTTP server on a random available port and returns the URL. Tell the user the URL so they can open it in their browser. The server runs in the background and persists for the session.

### Keyboard Navigation

The web UI is keyboard-native:

**Dashboard (index):**
- `↑` / `↓` or `,` / `.` - Navigate activity feed
- `→` or `Enter` - Open selected workstream
- `/` - Open command palette (fuzzy search)
- `g h` - Go home
- `?` - Toggle help modal
- `r` - Refresh

**Workstream detail:**
- `↑` / `↓` or `,` / `.` - Navigate tasks/logs
- `→` - Expand collapsed log
- `←` - Collapse log (or go back if not expandable)
- `Backspace` - Return to dashboard
- `/` - Open command palette
- `?` - Toggle help modal

## Exporting Workstreams

Export workstreams to markdown files for version control:

```bash
# Single workstream to stdout
streamctl export fleetadm/auth

# All workstreams for a project to a directory
streamctl export fleetadm --dir ./workstreams/
```

Exported files include a header marking them as generated. Use in pre-commit hooks:

```bash
#!/bin/bash
streamctl export myproject --dir ./workstreams/
git add workstreams/*.md
```

## needs_help vs blocked

- **blocked** = structural dependency, can't proceed until another workstream completes
- **needs_help** = signal for attention, workstream is stuck and needs human intervention

A workstream can be both. Use `needs_help=true` when repeatedly hitting issues.

## Claude Code Hooks

streamctl ships with hooks for Claude Code that auto-log progress.

### TaskCompleted Hook

Automatically logs task completions to the active workstream.

Add to `~/.claude/settings.json`:
```json
{
  "hooks": {
    "TaskCompleted": [{
      "hooks": [{
        "type": "command",
        "command": "/path/to/streamctl/hooks/task-completed.sh"
      }]
    }]
  }
}
```

When Claude marks a task complete via `TaskUpdate`, the hook:
1. Extracts the task subject
2. Detects the project from current directory
3. Logs "Completed: {task}" to the active (in_progress) workstream

### CLI Log Command

Log directly from the command line:

```bash
streamctl log myproject "Completed: implement auth"
```

Logs to the most recently updated `in_progress` workstream for that project.
