# streamctl

Workstream manager for coordinating parallel Claude Code agents.

## Quick Setup for Claude Code

**If you're a Claude Code agent reading this, follow these steps:**

1. **Build the binary:**
   ```bash
   cd ~/streamctl  # or wherever this repo is cloned
   go build -o streamctl ./cmd/streamctl
   ```

2. **Initialize the database:**
   ```bash
   ./streamctl init
   ```

3. **Add to Claude Code MCP settings** (`~/.claude/settings.json`):
   ```json
   {
     "mcpServers": {
       "streamctl": {
         "command": "/home/USER/streamctl/streamctl",
         "args": ["serve"]
       }
     }
   }
   ```
   Replace `/home/USER/streamctl/streamctl` with the actual path to the binary.

4. **Restart Claude Code** to load the MCP server.

After setup, you'll have these MCP tools available:

| Tool | Usage |
|------|-------|
| `workstream_list` | List workstreams. Filter with `project`, `state`, `owner` params. |
| `workstream_get` | Get full details. Requires `project` and `name`. |
| `workstream_create` | Create new. Requires `project`, `name`, `objective`. |
| `workstream_update` | Update fields: `state`, `log_entry`, `plan_index`. |
| `workstream_claim` | Set `owner` on a workstream. |
| `workstream_release` | Clear owner. |

---

## What is a Workstream?

A workstream tracks a unit of work:
- **project** - Repository/project name
- **name** - Descriptive name
- **state** - `pending`, `in_progress`, `blocked`, `done`
- **owner** - Which agent owns this (for coordination)
- **objective** - One-sentence goal
- **plan** - Checklist of steps
- **log** - Progress entries

## Typical Workflow

```
1. workstream_list                    # See available work
2. workstream_claim (project, name, owner)   # Claim it
3. workstream_update (log_entry)      # Log progress as you work
4. workstream_update (plan_index)     # Check off plan items
5. workstream_update (state=done)     # Mark complete
6. workstream_release                 # Release ownership
```

## Database Location

Priority order:
1. `STREAMCTL_DB` environment variable
2. `.streamctl/workstreams.db` in current directory (project-local)
3. `~/.streamctl/workstreams.db` (user global)

## CLI Commands

```bash
streamctl init     # Initialize database
streamctl serve    # Start MCP server (stdio) - used by Claude Code
streamctl tui      # Launch TUI dashboard
streamctl list     # List workstreams as JSON
streamctl help     # Show help
```

## TUI Dashboard

Run `./streamctl tui` for a visual overview:
- `j/k` - Navigate
- `Enter` - Select
- `Esc` - Back
- `q` - Quit
