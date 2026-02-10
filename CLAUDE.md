# streamctl

MCP server + TUI for managing workstreams across projects.

## Build Commands

```bash
make build                    # Build binary
make test                     # Run tests
./streamctl init              # Initialize database
./streamctl serve             # Start MCP server
./streamctl tui               # Launch TUI dashboard
./streamctl list              # List workstreams (JSON)
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
│   └── tui/                  # Terminal UI (bubbletea)
└── pkg/
    └── workstream/           # Core types and rendering
```

## Database Schema

```sql
workstreams (id, project, name, state, owner, objective, key_context, decisions, last_update, created_at)
plan_items (id, workstream_id, position, text, complete)
log_entries (id, workstream_id, timestamp, content)
```

## MCP Tools

| Tool | Description |
|------|-------------|
| `workstream_list` | List all workstreams, optionally filter by project/state/owner |
| `workstream_get` | Get full workstream content (rendered as markdown) |
| `workstream_create` | Create new workstream |
| `workstream_update` | Update state, add log entry, toggle plan item |
| `workstream_claim` | Set ownership of a workstream |
| `workstream_release` | Release ownership |

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
