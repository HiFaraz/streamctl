# streamctl

MCP server + TUI for managing workstreams across projects.

## Build Commands

```bash
make build                    # Build binary
make test                     # Run tests
./streamctl init              # Initialize database
./streamctl serve             # Start MCP server
./streamctl tui               # Launch TUI dashboard
./streamctl web               # Start web UI (auto-detects project)
./streamctl list              # List workstreams (JSON)
./streamctl export PROJECT/NAME          # Export single workstream to stdout
./streamctl export PROJECT --dir ./dir/  # Export all to directory
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
workstreams (id, project, name, state, owner, objective, key_context, decisions, needs_help, last_update, created_at)
plan_items (id, workstream_id, position, text, complete, status, notes)
log_entries (id, workstream_id, timestamp, content)
workstream_dependencies (blocker_id, blocked_id, created_at)
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

### workstream_update Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `project` | string | Project name (required) |
| `name` | string | Workstream name (required) |
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
- `.` / `,` - Navigate activity feed (down/up)
- `Enter` - Open selected workstream
- `/` - Open command palette (fuzzy search)
- `Backspace` - Return to dashboard
- `g h` - Go home
- `?` - Toggle help modal
- `r` - Refresh

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
