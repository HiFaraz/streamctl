# streamctl Changelog for LLMs

This file documents API changes for AI agents using streamctl MCP tools.

---

## Action Required for Existing Users

If you're already using streamctl, do the following:

### 1. Update Your Root CLAUDE.md

Add these new MCP tool parameters to your `~/.claude/CLAUDE.md` workstream documentation:

```markdown
### workstream_update Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `task_add` | string | Add new task with this text |
| `task_remove` | number | Remove task at position (0-indexed) |
| `task_status` | object | Set status: `{"position": 0, "status": "done"}` |
| `task_notes` | object | Set notes (markdown): `{"position": 0, "notes": "..."}` |
| `add_blocker` | string | Add dependency: `project/name` blocks this workstream |
| `remove_blocker` | string | Remove dependency |

**Task statuses:** `pending`, `in_progress`, `done`, `skipped`
**Task notes:** Supports markdown (code blocks, lists, links, headers)
**Log entries:** Also support markdown - use for rich context, code snippets, decision rationale
```

### 2. Rebuild and Restart MCP Server

The database schema auto-migrates on startup. You must:

```bash
cd ~/streamctl
make restart
```

The migration automatically:
- Adds `status` column to plan_items
- Converts `complete=true` to `status=done`
- Converts `complete=false` to `status=pending`
- Creates `workstream_dependencies` table

### 3. Update Your Permissions

Add these MCP commands to your allowed permissions in `~/.claude/settings.json` or approve them when prompted:

```
mcp__streamctl__workstream_update (with task_add, task_remove, task_status, add_blocker, remove_blocker)
```

### 4. Backfill Your Existing Workstreams

For each active workstream, update task statuses and add dependencies:

```
# Check current state
workstream_get(project="X", name="Y")

# Update task statuses (0-indexed positions)
workstream_update(project="X", name="Y", task_status={"position": 0, "status": "in_progress"})
workstream_update(project="X", name="Y", task_status={"position": 1, "status": "done"})

# Add dependencies if workstreams block each other
workstream_update(project="X", name="downstream", add_blocker="X/upstream")
```

### 5. Adopt New Workflow

- When starting a task: set status to `in_progress`
- When completing a task: set status to `done`
- When a task becomes irrelevant: set status to `skipped`
- When creating dependent workstreams: use `add_blocker`

---

## 2026-02-10: Tasks and Dependencies

### New Features

#### Task Management
Plan items now support status tracking beyond simple completion:

**Status values:** `pending`, `in_progress`, `done`, `skipped`

**New parameters for `workstream_update`:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `task_add` | string | Add a new task with this text |
| `task_remove` | number | Remove task at this position (0-indexed) |
| `task_status` | object | Set task status: `{"position": 0, "status": "done"}` |
| `task_notes` | object | Set task notes (markdown): `{"position": 0, "notes": "..."}` |

**Examples:**
```
# Add a task
workstream_update(project="myapp", name="auth", task_add="Implement JWT validation")

# Set task status
workstream_update(project="myapp", name="auth", task_status={"position": 0, "status": "in_progress"})

# Mark task done
workstream_update(project="myapp", name="auth", task_status={"position": 0, "status": "done"})

# Skip a task
workstream_update(project="myapp", name="auth", task_status={"position": 1, "status": "skipped"})

# Remove a task
workstream_update(project="myapp", name="auth", task_remove=2)

# Add notes to a task (supports markdown: code blocks, lists, links)
workstream_update(project="myapp", name="auth", task_notes={"position": 0, "notes": "## Details\n- Use RS256 algorithm\n- Token expiry: 1 hour\n\n```go\ntoken := jwt.New()\n```"})
```

#### Structured Dependencies
Track blocking relationships between workstreams:

**New parameters for `workstream_update`:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `add_blocker` | string | Add dependency: `project/workstream` blocks this one |
| `remove_blocker` | string | Remove dependency from this workstream |

**Examples:**
```
# auth blocks api (api depends on auth)
workstream_update(project="myapp", name="api", add_blocker="myapp/auth")

# Remove the dependency
workstream_update(project="myapp", name="api", remove_blocker="myapp/auth")
```

### Rendered Output Changes

`workstream_get` now shows:

**Task status markers:**
- `[ ]` - pending
- `[>]` - in_progress
- `[x]` - done
- `[-]` - skipped

**Dependencies section** (when present):
```markdown
## Dependencies
Blocked by:
- project/workstream-name

Blocks:
- project/downstream
```

### Migration Notes

- Existing plan items with `complete=true` are now `status=done`
- Existing plan items with `complete=false` are now `status=pending`
- The `plan_index` parameter still works for toggling completion (backward compatible)
- Dependencies start empty for all existing workstreams

### Recommended Usage Patterns

**When starting a task:**
```
workstream_update(project="X", name="Y", task_status={"position": 0, "status": "in_progress"})
```

**When completing a task:**
```
workstream_update(project="X", name="Y", task_status={"position": 0, "status": "done"})
```

**When a workstream depends on another:**
```
# Before starting work on "api", ensure "auth" is done
workstream_update(project="myapp", name="api", add_blocker="myapp/auth")
```

**When unblocked:**
```
# Auth is done, remove the blocker
workstream_update(project="myapp", name="api", remove_blocker="myapp/auth")
```
