# Changelog

## Unreleased

### Added

- **Milestone deletion**: `milestone_delete(project, name)` removes a milestone
  - Workstreams are NOT deleted - milestones are groupings that reference workstreams, not owners
  - Documentation clarified to explain the milestone-workstream relationship

- **Workstream renaming**: `workstream_update(project, name, new_name="...")` to rename a workstream

- **Markdown rendering in web UI logs**: Log entries now render markdown properly with line breaks

- **Collapsible logs in web UI**: Long logs (>3 lines or >200 chars) are collapsed by default
  - `→` to expand, `←` to collapse
  - Visual hint shows current state

- **Arrow key navigation in web UI**:
  - Dashboard: `→` or `Enter` opens selected item
  - Detail page: `→` expands log, `←` collapses (or goes back if not expandable)

### Removed

- **Removed `key_context` and `decisions` fields**: These were unused legacy fields
  - Use `objective` for context (description softened to allow longer content)
  - Use log entries to record decisions with timestamps

### Changed

- **Objective field description**: Changed from "One-sentence objective" to "Objective and context for this workstream" to allow richer content at creation time

- **Cross-workstream milestones**: Define gates that require multiple workstreams to complete
  - `milestone_create(project, name, description?)` - Create a milestone
  - `milestone_get(project, name)` - Get milestone with computed status and requirements list
  - `milestone_list(project?)` - List milestones with computed status
  - `milestone_update(project, name, add_requirement?, remove_requirement?, description?)` - Modify milestone
  - Status computed automatically: `pending` (none done), `in_progress` (some done), `done` (all done)
  - Ideal for coordinating agent teams: "Is wave 1 done?" via MCP without reading external files
  - Example:
    ```
    milestone_create(project="fleetadm", name="wave-1", description="Foundation layer")
    milestone_update(project="fleetadm", name="wave-1", add_requirement="fleetadm/auth")
    milestone_update(project="fleetadm", name="wave-1", add_requirement="fleetadm/api")
    milestone_get(project="fleetadm", name="wave-1")  # Shows status=in_progress if some done
    ```

- **Web UI server** (`web_serve` MCP tool): Start a browser-based dashboard to view workstreams
  - Modern dark theme with sidebar navigation
  - Dashboard home page with insights and activity feed
  - Auto-assigns a floating port (no conflicts)
  - Returns URL directly to Claude for sharing with the user
  - Usage: `web_serve(project="myproject")` returns `http://localhost:<port>`

- **Dashboard insights**: Home page shows at-a-glance status
  - "Needs Help" count (red highlight when > 0)
  - "Blocked" count (amber highlight when > 0)
  - "In Progress" count with quick links

- **Activity feed**: Recent log entries across all workstreams, newest first

- **Live dashboard**: Auto-refreshes every 5 seconds with smooth updates
  - Green "Live" indicator pulses to show active monitoring
  - Preserves scroll position during updates
  - Ideal for monitoring parallel agents/teams

- **Needs Help flag**: Mark workstreams as at-risk/stuck
  - `workstream_update(project, name, needs_help=true)` to flag
  - Displayed with `!` indicator in sidebar and lists
  - Surfaced prominently in dashboard insights

- **CLI web command**: `streamctl web [--port PORT]` starts the web UI from the terminal
  - Auto-detects project from git remote or directory name
  - Defaults to port 8080

- **CLI export command**: Export workstreams to markdown files for version control
  - `streamctl export PROJECT/NAME` - export single workstream to stdout
  - `streamctl export PROJECT --dir ./workstreams/` - export all to directory
  - Generated files include header warning not to edit manually
  - Ideal for pre-commit hooks to sync exported files

### Changed

- **Keyboard-native navigation**: Feed-centric UX where keyboard is the primary interaction
  - `.`/`,` to navigate through activity feed entries (down/up)
  - `Enter` to open the selected entry's workstream
  - `/` to open command palette for fuzzy search/jump to any workstream
  - `g h` to go home (dashboard)
  - `Backspace` to return to dashboard from detail view
  - `?` to toggle help modal with all shortcuts
  - `r` to manually refresh
  - Status bar shows available shortcuts at bottom of screen

- **WCAG 2.2 AAA compliant UI**: Complete redesign for accessibility
  - High contrast light theme (7:1+ contrast ratios)
  - Skip links for keyboard navigation
  - Proper ARIA landmarks and labels
  - Visible focus indicators (3px solid outline)
  - Semantic HTML with proper heading hierarchy
  - Screen reader announcements for live updates

- **Log entries now show newest first** (was oldest first)

- Updated CLAUDE.md with web UI documentation

## 1.0.0

### Added

- Core workstream management with SQLite storage
- MCP tools: `workstream_list`, `workstream_get`, `workstream_create`, `workstream_update`, `workstream_claim`, `workstream_release`
- Task management: add, remove, set status (pending/in_progress/done/skipped), add notes (markdown)
- Workstream dependencies: `add_blocker`, `remove_blocker`
- TUI dashboard with keyboard navigation
- Log entries with markdown support
