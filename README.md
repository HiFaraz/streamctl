# streamctl

Workstream manager for coordinating parallel Claude Code agents.

## For Claude Code Agents

**To set up streamctl in this project, run these commands:**

```bash
# Clone and build (skip if ~/streamctl exists)
if [ ! -f ~/streamctl/streamctl ]; then
  git clone https://github.com/HiFaraz/streamctl ~/streamctl
  cd ~/streamctl && go build -o streamctl ./cmd/streamctl
fi

# Initialize project-local database
mkdir -p .streamctl
~/streamctl/streamctl init

# Add to Claude Code MCP settings
# Edit ~/.claude/settings.json and add:
```

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

After adding to settings, tell the user to restart Claude Code.

---

## MCP Tools Available

| Tool | Parameters | Description |
|------|------------|-------------|
| `workstream_list` | `project?`, `state?`, `owner?` | List workstreams |
| `workstream_get` | `project`, `name` | Get full workstream as markdown |
| `workstream_create` | `project`, `name`, `objective` | Create workstream |
| `workstream_update` | `project`, `name`, `state?`, `log_entry?`, `plan_index?` | Update workstream |
| `workstream_claim` | `project`, `name`, `owner` | Claim ownership |
| `workstream_release` | `project`, `name` | Release ownership |

## Workflow

```
workstream_create(project, name, objective)  # Create
workstream_claim(project, name, owner)       # Claim before working
workstream_update(log_entry="...")           # Log progress
workstream_update(plan_index=0)              # Toggle plan item
workstream_update(state="done")              # Mark done
workstream_release(project, name)            # Release
```

## States

`pending` → `in_progress` → `done` (or `blocked`)

## Database

Uses `.streamctl/workstreams.db` in the current working directory. Add `.streamctl/` to `.gitignore`.
