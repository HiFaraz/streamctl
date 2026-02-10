# streamctl

A workstream management tool for coordinating parallel Claude Code agents across projects.

## What is a Workstream?

A workstream represents a unit of work with:
- **Project** - The repository or project it belongs to
- **Name** - A descriptive name for the work
- **State** - `pending`, `in_progress`, `blocked`, or `done`
- **Owner** - Which agent currently owns this work
- **Objective** - One-sentence goal
- **Plan** - Checklist of steps
- **Log** - Timestamped progress entries

## Installation

```bash
# Build from source
git clone https://github.com/faraz/streamctl
cd streamctl
make build

# Initialize the database
./streamctl init

# Optionally install to ~/bin
make install
```

## Claude Code MCP Setup

Add streamctl as an MCP server in your Claude Code settings:

**~/.claude/settings.json:**
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

After restarting Claude Code, you'll have access to these tools:

| Tool | Description |
|------|-------------|
| `workstream_list` | List workstreams (filter by project, state, owner) |
| `workstream_get` | Get full workstream details as markdown |
| `workstream_create` | Create a new workstream |
| `workstream_update` | Update state, add log entry, toggle plan items |
| `workstream_claim` | Set yourself as owner |
| `workstream_release` | Release ownership |

## Usage with Claude Code

### Starting a Session

When you start working on a project, check for existing workstreams:

```
Use workstream_list to see what work is available for project "myproject"
```

### Claiming Work

Before starting work, claim a workstream to prevent conflicts:

```
Use workstream_claim to take ownership of "feature-x" in project "myproject"
```

### Logging Progress

As you work, add log entries to track progress:

```
Use workstream_update to add a log entry: "Completed API endpoint, moving to tests"
```

### Updating Plan Items

Toggle plan items as you complete them:

```
Use workstream_update to toggle plan item 0 (first item) as complete
```

### Finishing Work

When done, update the state and release ownership:

```
Use workstream_update to set state to "done"
Use workstream_release to clear ownership
```

## TUI Dashboard

For a visual overview of all workstreams:

```bash
./streamctl tui
```

**Navigation:**
- `j/k` or arrows - Move cursor
- `Enter` - Select/drill down
- `Esc` - Go back
- `r` - Refresh
- `q` - Quit

## CLI Commands

```bash
streamctl init               # Initialize database
streamctl serve              # Start MCP server (stdio)
streamctl tui                # Launch TUI dashboard
streamctl list               # List all workstreams as JSON
streamctl list --project X   # Filter by project
streamctl version            # Show version
streamctl help               # Show help
```

## Database Location

Each user runs their own streamctl instance. Database location priority:

1. `STREAMCTL_DB` environment variable (if set)
2. `.streamctl/workstreams.db` in current directory (project-local)
3. `~/.streamctl/workstreams.db` (user global)

**For project-local databases:**
```bash
mkdir .streamctl
streamctl init  # Creates .streamctl/workstreams.db in cwd
```

**For user-global database:**
```bash
streamctl init  # Creates ~/.streamctl/workstreams.db
```

Add `.streamctl/` to your `.gitignore` if using project-local databases.

## Example Workflow

1. **Plan parallel work** - Create workstreams for independent tasks
2. **Assign to agents** - Each agent claims a workstream
3. **Track progress** - Agents log their progress
4. **Coordinate** - Use the TUI to see overall status
5. **Complete** - Mark workstreams as done when finished

## CLAUDE.md Integration

Add to your project's CLAUDE.md:

```markdown
## Workstreams

This project uses streamctl for parallel work coordination.

Before starting work:
1. Check `workstream_list` for available tasks
2. Claim a workstream with `workstream_claim`
3. Log progress with `workstream_update`
4. Mark complete and release when done
```
