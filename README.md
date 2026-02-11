# streamctl

Persistent workstream tracking for Claude Code agents.

## The Problem

When running Claude Code agents across multiple sessions - whether solo or in teams - context is lost:
- What was I working on yesterday?
- What decisions did I make and why?
- Which pieces are blocked and on what?
- Where did the last session leave off?

Agent teams solve in-session coordination beautifully (task lists, mailbox, claiming). But when the session ends, everything disappears.

## The Solution

streamctl provides a persistent database of **workstreams** - units of work that survive across sessions. It's long-term memory for your agent work.

```
workstream_create(project="myapp", name="auth", objective="Implement JWT authentication")
workstream_update(project="myapp", name="auth", log_entry="Completed token validation. Next: middleware.")
workstream_update(project="myapp", name="auth", state="done")
```

Next week, a new session can read exactly where things stand:

```
workstream_list(project="myapp")
workstream_get(project="myapp", name="auth")  # Full history, decisions, context
```

---

## Examples

### Parallel Feature Development

Break a feature into independent workstreams:

```
workstream_create(project="myapp", name="db-schema", objective="Add user preferences table")
workstream_create(project="myapp", name="api", objective="CRUD endpoints for preferences")
workstream_create(project="myapp", name="frontend", objective="Preferences settings page")
```

Multiple agents (or team sessions) each claim one, work on it, and log progress. No conflicts.

### Resuming Work

Yesterday's session ended mid-feature:

```
workstream_list(project="myapp", state="in_progress")
→ Shows auth workstream with log: "JWT validation done. Next: middleware integration."
```

Pick up exactly where you left off.

### Tracking Blockers

```
workstream_update(project="myapp", name="api", state="blocked",
  log_entry="Waiting for auth workstream to finish token format")
```

Tomorrow's session sees this is blocked and why.

---

## Using with Agent Teams

If you're using [Claude Code agent teams](https://code.claude.com/docs/en/agent-teams), streamctl adds the persistence layer on top.

**The mental model:**
- **Workstreams** = epics that span multiple sessions (persistent)
- **Team tasks** = granular work within a session (ephemeral)
- **Mailbox** = real-time chatter during a session (ephemeral)
- **streamctl logs** = decisions and milestones (persistent)

| Ephemeral (agent teams) | Persistent (streamctl) |
|-------------------------|------------------------|
| Task list | Workstreams |
| Mailbox messages | Log entries |
| Team context | Objective, key context, decisions |

### Pattern for Team Leads

**Starting a session:**
```
1. workstream_list(project="myapp")           # What needs work?
2. workstream_get(project, name) for each     # Read previous context
3. Create team, passing workstream context to teammates
```

**During the session:**
- Teammates use mailbox for quick coordination ("are you done with X?")
- Lead logs major milestones to streamctl (decisions, blockers, completions)
- Don't log every small step - workstreams are higher-level than tasks

**Ending a session:**
```
workstream_update(project="myapp", name="auth",
  state="in_progress",
  log_entry="JWT done. Middleware 50%. Decision: using httpOnly cookies. Next: finish middleware, then integrate with api workstream.")
```

**Key principles:**
- Lead owns streamctl updates (teammates focus on work)
- Log decisions with rationale ("chose X because Y")
- Write for your future self - assume no memory
- Update states honestly - blocked means blocked

---

## Setup

```bash
# Build (one-time)
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

Restart Claude Code. Add `.streamctl/` to `.gitignore`.

---

## Quick Start Prompt

Paste this into Claude Code to set up streamctl:

```
Read ~/streamctl/README.md, then:

1. Add streamctl MCP server to ~/.claude/settings.json if not already configured
2. Add to ~/.claude/CLAUDE.md (create if needed) instructions to use workstreams:
   - At session start: check workstream_list, read context with workstream_get for in_progress work
   - During work: create workstreams for non-trivial features, log progress and decisions
   - At session end: update state and log what's next
3. Initialize .streamctl/ in the current project and add to .gitignore
4. Check for existing workstreams and resume any in_progress work
```

---

## Reference

### MCP Tools

| Tool | Parameters | Description |
|------|------------|-------------|
| `workstream_list` | `project?`, `state?`, `owner?` | List workstreams |
| `workstream_get` | `project`, `name` | Get full details |
| `workstream_create` | `project`, `name`, `objective` | Create workstream |
| `workstream_update` | `project`, `name`, `state?`, `log_entry?`, `plan_index?` | Update |
| `workstream_claim` | `project`, `name`, `owner` | Set owner |
| `workstream_release` | `project`, `name` | Clear owner |

### Workstream Fields

| Field | Description |
|-------|-------------|
| `project` | Repository or project name |
| `name` | Identifier (e.g., "auth-refactor") |
| `state` | `pending` → `in_progress` → `done` / `blocked` |
| `owner` | Who's working on it |
| `objective` | One-sentence goal |
| `plan` | Checklist (toggle items with `plan_index`) |
| `log` | Timestamped progress entries |

### CLI

```bash
streamctl tui      # Visual dashboard
streamctl list     # JSON dump
streamctl help     # Help
```
