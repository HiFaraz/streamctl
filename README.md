# streamctl

Coordinate parallel Claude Code agents working on the same codebase.

## The Problem

When running multiple Claude Code agents in parallel (e.g., one working on auth, another on the API, another on tests), they have no way to:
- Know what work is available or claimed
- Avoid stepping on each other's toes
- Track progress across sessions
- Hand off work between sessions

## The Solution

streamctl provides a **persistent** database of **workstreams** - units of work that survive across sessions. Think of it as long-term memory for your agent coordination.

### How It Complements Agent Teams

[Claude Code agent teams](https://code.claude.com/docs/en/agent-teams) already provide excellent in-session coordination:
- **Shared task list** with claiming and dependencies
- **Mailbox** for direct messaging between teammates
- **Lead/teammate** hierarchy with delegation

But agent teams are **ephemeral** - when the team ends, everything disappears. This creates problems:
- Can't resume work the next day
- Decisions and context are lost
- No handoff between different team sessions
- `/resume` doesn't restore teammates

**streamctl adds the persistence layer:**

| Agent Teams (in-session) | streamctl (cross-session) |
|--------------------------|---------------------------|
| Task list disappears when team ends | Workstreams persist forever |
| Mailbox messages are ephemeral | Log entries are permanent |
| No memory of past decisions | Decisions section preserved |
| Can't resume teammates | Pick up where you left off |

### Recommended Workflow

**1. Before starting a team session:**
```
workstream_list(project="myapp")  # See what work exists
workstream_get(project="myapp", name="auth-refactor")  # Read context from last session
```

**2. Create the agent team based on workstreams:**
```
"Create an agent team. We have these workstreams to complete:
- auth-refactor: [paste objective and context from workstream_get]
- api-endpoints: [paste from another workstream]
Assign one teammate per workstream."
```

**3. During the session, teammates use the mailbox for quick coordination:**
```
Lead: "Auth teammate, are you blocked on anything?"
Auth: "Waiting for the User model changes from API teammate"
```

**4. At key milestones, log to streamctl for persistence:**
```
workstream_update(project="myapp", name="auth-refactor",
  log_entry="Implemented JWT validation. Blocked on User model - API teammate working on it.")
```

**5. When the team session ends, update states:**
```
workstream_update(project="myapp", name="auth-refactor", state="blocked",
  log_entry="Session ended. JWT done, waiting on User model. Next: integrate with API.")
workstream_update(project="myapp", name="api-endpoints", state="in_progress",
  log_entry="User model 80% complete. Continuing tomorrow.")
```

**6. Next day, new team session picks up with full context:**
```
workstream_list(project="myapp", state="in_progress")
# Shows exactly where each piece of work stands
```

### When to Use What

| Use Case | Tool |
|----------|------|
| Quick coordination during a session | Agent team mailbox |
| Claiming tasks within a session | Agent team task list |
| Preserving context across sessions | streamctl workstreams |
| Recording decisions for future reference | streamctl log entries |
| Resuming work days later | streamctl |
| Parallel exploration (research, debugging) | Agent teams |
| Long-running projects with multiple sessions | streamctl + agent teams |

### Recommended Pattern for Agent Team Leads

**At session start:**
```
1. workstream_list(project="myapp")
   → See all workstreams and their states

2. For each in_progress or pending workstream:
   workstream_get(project="myapp", name="...")
   → Read objective, key context, decisions, and logs from previous sessions

3. Create agent team with context:
   "Create a team to work on these workstreams:
   - auth-refactor: [paste objective + key context + recent log]
   - api-endpoints: [paste objective + where we left off]
   Spawn one teammate per workstream."
```

**During session:**
```
• Teammates coordinate via mailbox (ephemeral, real-time)
• Lead logs significant milestones to streamctl:
  - Major decisions with rationale
  - Completed milestones
  - Blockers discovered

• Don't log every small step - workstreams are higher-level than tasks
```

**At session end:**
```
For each workstream touched:
  workstream_update(project="myapp", name="auth-refactor",
    state="in_progress",  # or "done" or "blocked"
    log_entry="Completed JWT validation with RS256. Middleware done.
               Next: integrate with User model from api-endpoints.
               Decision: Using httpOnly cookies for token storage.")
```

**Key principles:**
- **Lead owns streamctl updates** - teammates focus on their work
- **Log decisions, not just actions** - "chose X because Y" helps future sessions
- **Write for your future self** - assume no memory of this session
- **Update states honestly** - blocked means blocked, don't leave things "in_progress" if stuck

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
