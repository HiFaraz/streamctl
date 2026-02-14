package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/faraz/streamctl/pkg/workstream"
)

// Filter for listing workstreams
type Filter struct {
	Project string
	State   workstream.State
	Owner   string
}

// WorkstreamUpdate for partial updates
type WorkstreamUpdate struct {
	State     *workstream.State
	Owner     *string
	LogEntry  *string // Append to log
	PlanIndex *int    // Toggle plan item completion
	NeedsHelp *bool   // Flag for at-risk/stuck workstreams
}

// Store provides SQLite-backed CRUD operations for workstreams
type Store struct {
	db *sql.DB
}

// New creates a new Store with the given database path
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates the database schema and runs migrations for existing databases
func (s *Store) migrate() error {
	// Base schema (for new databases)
	schema := `
	CREATE TABLE IF NOT EXISTS workstreams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project TEXT NOT NULL,
		name TEXT NOT NULL,
		state TEXT NOT NULL DEFAULT 'pending',
		owner TEXT DEFAULT '',
		needs_help BOOLEAN DEFAULT FALSE,
		objective TEXT DEFAULT '',
		key_context TEXT DEFAULT '',
		decisions TEXT DEFAULT '',
		last_update DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(project, name)
	);

	CREATE TABLE IF NOT EXISTS plan_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workstream_id INTEGER NOT NULL REFERENCES workstreams(id) ON DELETE CASCADE,
		position INTEGER NOT NULL,
		text TEXT NOT NULL,
		complete BOOLEAN DEFAULT FALSE,
		status TEXT NOT NULL DEFAULT 'pending',
		notes TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS log_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workstream_id INTEGER NOT NULL REFERENCES workstreams(id) ON DELETE CASCADE,
		timestamp DATETIME NOT NULL,
		content TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS workstream_dependencies (
		blocker_id INTEGER NOT NULL REFERENCES workstreams(id) ON DELETE CASCADE,
		blocked_id INTEGER NOT NULL REFERENCES workstreams(id) ON DELETE CASCADE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (blocker_id, blocked_id),
		CHECK(blocker_id != blocked_id)
	);

	CREATE INDEX IF NOT EXISTS idx_workstreams_project ON workstreams(project);
	CREATE INDEX IF NOT EXISTS idx_workstreams_state ON workstreams(state);
	CREATE INDEX IF NOT EXISTS idx_workstreams_owner ON workstreams(owner);
	CREATE INDEX IF NOT EXISTS idx_plan_items_workstream ON plan_items(workstream_id);
	CREATE INDEX IF NOT EXISTS idx_log_entries_workstream ON log_entries(workstream_id);
	CREATE INDEX IF NOT EXISTS idx_deps_blocker ON workstream_dependencies(blocker_id);
	CREATE INDEX IF NOT EXISTS idx_deps_blocked ON workstream_dependencies(blocked_id);

	CREATE TABLE IF NOT EXISTS milestones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(project, name)
	);

	CREATE TABLE IF NOT EXISTS milestone_requirements (
		milestone_id INTEGER NOT NULL REFERENCES milestones(id) ON DELETE CASCADE,
		workstream_id INTEGER NOT NULL REFERENCES workstreams(id) ON DELETE CASCADE,
		PRIMARY KEY (milestone_id, workstream_id)
	);

	CREATE INDEX IF NOT EXISTS idx_milestones_project ON milestones(project);
	CREATE INDEX IF NOT EXISTS idx_milestone_reqs_milestone ON milestone_requirements(milestone_id);
	`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	// Migration: Add status column to plan_items if missing
	if !s.columnExists("plan_items", "status") {
		if _, err := s.db.Exec(`ALTER TABLE plan_items ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'`); err != nil {
			return err
		}
		// Migrate existing data: complete=true -> done, complete=false -> pending
		if _, err := s.db.Exec(`UPDATE plan_items SET status = CASE WHEN complete THEN 'done' ELSE 'pending' END`); err != nil {
			return err
		}
	}

	// Migration: Add notes column to plan_items if missing
	if !s.columnExists("plan_items", "notes") {
		if _, err := s.db.Exec(`ALTER TABLE plan_items ADD COLUMN notes TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}

	// Migration: Add needs_help column to workstreams if missing
	if !s.columnExists("workstreams", "needs_help") {
		if _, err := s.db.Exec(`ALTER TABLE workstreams ADD COLUMN needs_help BOOLEAN DEFAULT FALSE`); err != nil {
			return err
		}
	}

	return nil
}

// columnExists checks if a column exists in a table
func (s *Store) columnExists(table, column string) bool {
	rows, err := s.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false
		}
		if name == column {
			return true
		}
	}
	return false
}

// Create creates a new workstream
func (s *Store) Create(ws *workstream.Workstream) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert workstream
	result, err := tx.Exec(`
		INSERT INTO workstreams (project, name, state, owner, objective, last_update)
		VALUES (?, ?, ?, ?, ?, ?)`,
		ws.Project, ws.Name, string(ws.State), ws.Owner, ws.Objective, ws.LastUpdate,
	)
	if err != nil {
		return err
	}

	wsID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Insert plan items
	for i, item := range ws.Plan {
		status := item.Status
		if status == "" {
			if item.Complete {
				status = workstream.TaskDone
			} else {
				status = workstream.TaskPending
			}
		}
		_, err := tx.Exec(`
			INSERT INTO plan_items (workstream_id, position, text, complete, status)
			VALUES (?, ?, ?, ?, ?)`,
			wsID, i, item.Text, item.Complete, string(status),
		)
		if err != nil {
			return err
		}
	}

	// Insert log entries
	for _, entry := range ws.Log {
		_, err := tx.Exec(`
			INSERT INTO log_entries (workstream_id, timestamp, content)
			VALUES (?, ?, ?)`,
			wsID, entry.Timestamp, entry.Content,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Get retrieves a workstream by project and name
func (s *Store) Get(project, name string) (*workstream.Workstream, error) {
	ws := &workstream.Workstream{}
	var wsID int64

	err := s.db.QueryRow(`
		SELECT id, project, name, state, owner, needs_help, objective, last_update
		FROM workstreams WHERE project = ? AND name = ?`,
		project, name,
	).Scan(&wsID, &ws.Project, &ws.Name, &ws.State, &ws.Owner, &ws.NeedsHelp, &ws.Objective, &ws.LastUpdate)
	if err != nil {
		return nil, err
	}

	// Load plan items
	rows, err := s.db.Query(`
		SELECT text, complete, status, notes FROM plan_items
		WHERE workstream_id = ? ORDER BY position`,
		wsID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item workstream.PlanItem
		if err := rows.Scan(&item.Text, &item.Complete, &item.Status, &item.Notes); err != nil {
			return nil, err
		}
		ws.Plan = append(ws.Plan, item)
	}

	// Load log entries
	logRows, err := s.db.Query(`
		SELECT timestamp, content FROM log_entries
		WHERE workstream_id = ? ORDER BY timestamp DESC`,
		wsID,
	)
	if err != nil {
		return nil, err
	}
	defer logRows.Close()

	for logRows.Next() {
		var entry workstream.LogEntry
		if err := logRows.Scan(&entry.Timestamp, &entry.Content); err != nil {
			return nil, err
		}
		ws.Log = append(ws.Log, entry)
	}

	// Load BlockedBy (workstreams that block this one)
	blockedByRows, err := s.db.Query(`
		SELECT w.project, w.name FROM workstream_dependencies d
		JOIN workstreams w ON d.blocker_id = w.id
		WHERE d.blocked_id = ?`, wsID)
	if err != nil {
		return nil, err
	}
	defer blockedByRows.Close()

	for blockedByRows.Next() {
		var dep workstream.Dependency
		if err := blockedByRows.Scan(&dep.BlockerProject, &dep.BlockerName); err != nil {
			return nil, err
		}
		dep.BlockedProject = ws.Project
		dep.BlockedName = ws.Name
		ws.BlockedBy = append(ws.BlockedBy, dep)
	}

	// Load Blocks (workstreams this one blocks)
	blocksRows, err := s.db.Query(`
		SELECT w.project, w.name FROM workstream_dependencies d
		JOIN workstreams w ON d.blocked_id = w.id
		WHERE d.blocker_id = ?`, wsID)
	if err != nil {
		return nil, err
	}
	defer blocksRows.Close()

	for blocksRows.Next() {
		var dep workstream.Dependency
		if err := blocksRows.Scan(&dep.BlockedProject, &dep.BlockedName); err != nil {
			return nil, err
		}
		dep.BlockerProject = ws.Project
		dep.BlockerName = ws.Name
		ws.Blocks = append(ws.Blocks, dep)
	}

	return ws, nil
}

// ListProjects returns all distinct project names
func (s *Store) ListProjects() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT project FROM workstreams ORDER BY project`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// List returns workstreams matching the filter
func (s *Store) List(filter Filter) ([]workstream.Workstream, error) {
	query := `SELECT id, project, name, state, owner, needs_help, objective, last_update FROM workstreams WHERE 1=1`
	var args []any

	if filter.Project != "" {
		query += " AND project = ?"
		args = append(args, filter.Project)
	}
	if filter.State != "" {
		query += " AND state = ?"
		args = append(args, string(filter.State))
	}
	if filter.Owner != "" {
		query += " AND owner = ?"
		args = append(args, filter.Owner)
	}

	query += " ORDER BY project, name"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []workstream.Workstream
	for rows.Next() {
		var ws workstream.Workstream
		var wsID int64
		if err := rows.Scan(&wsID, &ws.Project, &ws.Name, &ws.State, &ws.Owner, &ws.NeedsHelp, &ws.Objective, &ws.LastUpdate); err != nil {
			return nil, err
		}

		// Load plan items
		planRows, err := s.db.Query(`SELECT text, complete, status, notes FROM plan_items WHERE workstream_id = ? ORDER BY position`, wsID)
		if err != nil {
			return nil, err
		}
		for planRows.Next() {
			var item workstream.PlanItem
			planRows.Scan(&item.Text, &item.Complete, &item.Status, &item.Notes)
			ws.Plan = append(ws.Plan, item)
		}
		planRows.Close()

		// Load log entries
		logRows, err := s.db.Query(`SELECT timestamp, content FROM log_entries WHERE workstream_id = ? ORDER BY timestamp DESC`, wsID)
		if err != nil {
			return nil, err
		}
		for logRows.Next() {
			var entry workstream.LogEntry
			logRows.Scan(&entry.Timestamp, &entry.Content)
			ws.Log = append(ws.Log, entry)
		}
		logRows.Close()

		results = append(results, ws)
	}

	return results, nil
}

// Update applies partial updates to a workstream
func (s *Store) Update(project, name string, updates WorkstreamUpdate) error {
	// Get the workstream ID
	var wsID int64
	err := s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, project, name).Scan(&wsID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update state
	if updates.State != nil {
		_, err := tx.Exec(`UPDATE workstreams SET state = ?, last_update = ? WHERE id = ?`,
			string(*updates.State), time.Now().UTC(), wsID)
		if err != nil {
			return err
		}
	}

	// Update owner
	if updates.Owner != nil {
		_, err := tx.Exec(`UPDATE workstreams SET owner = ?, last_update = ? WHERE id = ?`,
			*updates.Owner, time.Now().UTC(), wsID)
		if err != nil {
			return err
		}
	}

	// Update needs_help flag
	if updates.NeedsHelp != nil {
		_, err := tx.Exec(`UPDATE workstreams SET needs_help = ?, last_update = ? WHERE id = ?`,
			*updates.NeedsHelp, time.Now().UTC(), wsID)
		if err != nil {
			return err
		}
	}

	// Append log entry
	if updates.LogEntry != nil {
		// Unescape literal \n and \u000A to actual newlines (MCP sends escaped newlines)
		content := strings.ReplaceAll(*updates.LogEntry, "\\n", "\n")
		content = strings.ReplaceAll(content, "\\u000A", "\n")
		_, err := tx.Exec(`INSERT INTO log_entries (workstream_id, timestamp, content) VALUES (?, ?, ?)`,
			wsID, time.Now().UTC(), content)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE workstreams SET last_update = ? WHERE id = ?`, time.Now().UTC(), wsID)
		if err != nil {
			return err
		}
	}

	// Toggle plan item
	if updates.PlanIndex != nil {
		_, err := tx.Exec(`
			UPDATE plan_items SET complete = NOT complete
			WHERE workstream_id = ? AND position = ?`,
			wsID, *updates.PlanIndex)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE workstreams SET last_update = ? WHERE id = ?`, time.Now().UTC(), wsID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AddTask adds a new task to a workstream
func (s *Store) AddTask(project, name, text string) error {
	var wsID int64
	err := s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, project, name).Scan(&wsID)
	if err != nil {
		return err
	}

	// Get next position
	var maxPos sql.NullInt64
	s.db.QueryRow(`SELECT MAX(position) FROM plan_items WHERE workstream_id = ?`, wsID).Scan(&maxPos)
	nextPos := 0
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	_, err = s.db.Exec(`
		INSERT INTO plan_items (workstream_id, position, text, complete, status)
		VALUES (?, ?, ?, FALSE, 'pending')`,
		wsID, nextPos, text,
	)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`UPDATE workstreams SET last_update = ? WHERE id = ?`, time.Now().UTC(), wsID)
	return err
}

// RemoveTask removes a task at the given position and reorders remaining tasks
func (s *Store) RemoveTask(project, name string, position int) error {
	var wsID int64
	err := s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, project, name).Scan(&wsID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete the task at position
	_, err = tx.Exec(`DELETE FROM plan_items WHERE workstream_id = ? AND position = ?`, wsID, position)
	if err != nil {
		return err
	}

	// Reorder remaining tasks (decrement position for all tasks after the deleted one)
	_, err = tx.Exec(`UPDATE plan_items SET position = position - 1 WHERE workstream_id = ? AND position > ?`, wsID, position)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE workstreams SET last_update = ? WHERE id = ?`, time.Now().UTC(), wsID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// SetTaskStatus sets the status of a task at the given position
func (s *Store) SetTaskStatus(project, name string, position int, status workstream.TaskStatus) error {
	var wsID int64
	err := s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, project, name).Scan(&wsID)
	if err != nil {
		return err
	}

	// Update status and complete (complete = true when status is done)
	complete := status == workstream.TaskDone
	_, err = s.db.Exec(`
		UPDATE plan_items SET status = ?, complete = ?
		WHERE workstream_id = ? AND position = ?`,
		string(status), complete, wsID, position,
	)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`UPDATE workstreams SET last_update = ? WHERE id = ?`, time.Now().UTC(), wsID)
	return err
}

// AddDependency creates a blocking relationship between two workstreams
func (s *Store) AddDependency(blockerProject, blockerName, blockedProject, blockedName string) error {
	var blockerID, blockedID int64

	err := s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, blockerProject, blockerName).Scan(&blockerID)
	if err != nil {
		return fmt.Errorf("blocker workstream not found: %s/%s", blockerProject, blockerName)
	}

	err = s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, blockedProject, blockedName).Scan(&blockedID)
	if err != nil {
		return fmt.Errorf("blocked workstream not found: %s/%s", blockedProject, blockedName)
	}

	_, err = s.db.Exec(`INSERT INTO workstream_dependencies (blocker_id, blocked_id) VALUES (?, ?)`, blockerID, blockedID)
	return err
}

// RemoveDependency removes a blocking relationship between two workstreams
func (s *Store) RemoveDependency(blockerProject, blockerName, blockedProject, blockedName string) error {
	var blockerID, blockedID int64

	err := s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, blockerProject, blockerName).Scan(&blockerID)
	if err != nil {
		return err
	}

	err = s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, blockedProject, blockedName).Scan(&blockedID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`DELETE FROM workstream_dependencies WHERE blocker_id = ? AND blocked_id = ?`, blockerID, blockedID)
	return err
}

// SetTaskNotes sets the notes for a task at the given position
func (s *Store) SetTaskNotes(project, name string, position int, notes string) error {
	var wsID int64
	err := s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, project, name).Scan(&wsID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`UPDATE plan_items SET notes = ? WHERE workstream_id = ? AND position = ?`,
		notes, wsID, position)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`UPDATE workstreams SET last_update = ? WHERE id = ?`, time.Now().UTC(), wsID)
	return err
}

// Rename renames a workstream
func (s *Store) Rename(project, oldName, newName string) error {
	result, err := s.db.Exec(`UPDATE workstreams SET name = ?, last_update = ? WHERE project = ? AND name = ?`,
		newName, time.Now().UTC(), project, oldName)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("workstream not found: %s/%s", project, oldName)
	}
	return nil
}

// Delete removes a workstream
func (s *Store) Delete(project, name string) error {
	result, err := s.db.Exec(`DELETE FROM workstreams WHERE project = ? AND name = ?`, project, name)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("workstream not found: %s/%s", project, name)
	}
	return nil
}

// RecentActivity returns recent log entries across all workstreams for a project
func (s *Store) RecentActivity(project string, limit, offset int) ([]workstream.ActivityEntry, error) {
	rows, err := s.db.Query(`
		SELECT w.name, w.project, l.timestamp, l.content, w.needs_help,
			(SELECT b.project || '/' || b.name
			 FROM workstream_dependencies d
			 JOIN workstreams b ON d.blocker_id = b.id
			 WHERE d.blocked_id = w.id
			 LIMIT 1)
		FROM log_entries l
		JOIN workstreams w ON l.workstream_id = w.id
		WHERE w.project = ?
		ORDER BY l.timestamp DESC
		LIMIT ? OFFSET ?`,
		project, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []workstream.ActivityEntry
	for rows.Next() {
		var entry workstream.ActivityEntry
		var blockedBy *string
		if err := rows.Scan(&entry.WorkstreamName, &entry.WorkstreamProject, &entry.Timestamp, &entry.Content, &entry.NeedsHelp, &blockedBy); err != nil {
			return nil, err
		}
		if blockedBy != nil {
			entry.BlockedBy = *blockedBy
		}
		entry.RelativeTime = relativeTime(entry.Timestamp)
		entries = append(entries, entry)
	}
	return entries, nil
}

// relativeTime returns a human-readable relative time string
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// SearchResult represents a search result (log entry or task)
type SearchResult struct {
	Type           string    `json:"type"` // "log" or "task"
	WorkstreamName string    `json:"workstreamName"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp,omitempty"`
	TaskPosition   int       `json:"taskPosition,omitempty"`
	TaskStatus     string    `json:"taskStatus,omitempty"`
	RelativeTime   string    `json:"relativeTime,omitempty"`
}

// Search searches logs and tasks for a query string
func (s *Store) Search(project, query, wsFilter string) ([]SearchResult, error) {
	if query == "" && wsFilter == "" {
		return []SearchResult{}, nil
	}

	var results []SearchResult

	// Search log entries
	logQuery := `
		SELECT w.name, l.timestamp, l.content
		FROM log_entries l
		JOIN workstreams w ON l.workstream_id = w.id
		WHERE w.project = ?`
	logArgs := []any{project}

	if query != "" {
		logQuery += " AND LOWER(l.content) LIKE LOWER(?)"
		logArgs = append(logArgs, "%"+query+"%")
	}
	if wsFilter != "" {
		// Case-insensitive partial match on workstream name
		logQuery += " AND LOWER(w.name) LIKE LOWER(?)"
		logArgs = append(logArgs, "%"+wsFilter+"%")
	}
	logQuery += " ORDER BY l.timestamp DESC LIMIT 50"

	rows, err := s.db.Query(logQuery, logArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.WorkstreamName, &r.Timestamp, &r.Content); err != nil {
			return nil, err
		}
		r.Type = "log"
		r.RelativeTime = relativeTime(r.Timestamp)
		results = append(results, r)
	}

	// Search tasks
	taskQuery := `
		SELECT w.name, p.position, p.text, p.status
		FROM plan_items p
		JOIN workstreams w ON p.workstream_id = w.id
		WHERE w.project = ?`
	taskArgs := []any{project}

	if query != "" {
		taskQuery += " AND (LOWER(p.text) LIKE LOWER(?) OR LOWER(p.notes) LIKE LOWER(?))"
		taskArgs = append(taskArgs, "%"+query+"%", "%"+query+"%")
	}
	if wsFilter != "" {
		// Case-insensitive partial match on workstream name
		taskQuery += " AND LOWER(w.name) LIKE LOWER(?)"
		taskArgs = append(taskArgs, "%"+wsFilter+"%")
	}
	taskQuery += " ORDER BY w.name, p.position LIMIT 50"

	taskRows, err := s.db.Query(taskQuery, taskArgs...)
	if err != nil {
		return nil, err
	}
	defer taskRows.Close()

	for taskRows.Next() {
		var r SearchResult
		if err := taskRows.Scan(&r.WorkstreamName, &r.TaskPosition, &r.Content, &r.TaskStatus); err != nil {
			return nil, err
		}
		r.Type = "task"
		results = append(results, r)
	}

	return results, nil
}

// CreateMilestone creates a new milestone
func (s *Store) CreateMilestone(m *workstream.Milestone) error {
	_, err := s.db.Exec(`
		INSERT INTO milestones (project, name, description)
		VALUES (?, ?, ?)`,
		m.Project, m.Name, m.Description,
	)
	return err
}

// GetMilestone retrieves a milestone by project and name
func (s *Store) GetMilestone(project, name string) (*workstream.Milestone, error) {
	m := &workstream.Milestone{}
	var milestoneID int64

	err := s.db.QueryRow(`
		SELECT id, project, name, description, created_at
		FROM milestones WHERE project = ? AND name = ?`,
		project, name,
	).Scan(&milestoneID, &m.Project, &m.Name, &m.Description, &m.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Load requirements with current workstream states
	rows, err := s.db.Query(`
		SELECT w.project, w.name, w.state
		FROM milestone_requirements mr
		JOIN workstreams w ON mr.workstream_id = w.id
		WHERE mr.milestone_id = ?`,
		milestoneID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var req workstream.MilestoneRequirement
		if err := rows.Scan(&req.WorkstreamProject, &req.WorkstreamName, &req.WorkstreamState); err != nil {
			return nil, err
		}
		m.Requirements = append(m.Requirements, req)
	}

	// Compute status based on requirements
	m.Status = computeMilestoneStatus(m.Requirements)

	return m, nil
}

// ListMilestones returns milestones, optionally filtered by project
func (s *Store) ListMilestones(project string) ([]workstream.Milestone, error) {
	query := `SELECT id, project, name, description, created_at FROM milestones`
	var args []any

	if project != "" {
		query += " WHERE project = ?"
		args = append(args, project)
	}
	query += " ORDER BY project, name"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var milestones []workstream.Milestone
	for rows.Next() {
		var m workstream.Milestone
		var milestoneID int64
		if err := rows.Scan(&milestoneID, &m.Project, &m.Name, &m.Description, &m.CreatedAt); err != nil {
			return nil, err
		}

		// Load requirements for each milestone
		reqRows, err := s.db.Query(`
			SELECT w.project, w.name, w.state
			FROM milestone_requirements mr
			JOIN workstreams w ON mr.workstream_id = w.id
			WHERE mr.milestone_id = ?`, milestoneID)
		if err != nil {
			return nil, err
		}
		for reqRows.Next() {
			var req workstream.MilestoneRequirement
			reqRows.Scan(&req.WorkstreamProject, &req.WorkstreamName, &req.WorkstreamState)
			m.Requirements = append(m.Requirements, req)
		}
		reqRows.Close()

		m.Status = computeMilestoneStatus(m.Requirements)
		milestones = append(milestones, m)
	}

	return milestones, nil
}

// AddMilestoneRequirement adds a workstream as a requirement for a milestone
func (s *Store) AddMilestoneRequirement(milestoneProject, milestoneName, wsProject, wsName string) error {
	var milestoneID, wsID int64

	err := s.db.QueryRow(`SELECT id FROM milestones WHERE project = ? AND name = ?`, milestoneProject, milestoneName).Scan(&milestoneID)
	if err != nil {
		return fmt.Errorf("milestone not found: %s/%s", milestoneProject, milestoneName)
	}

	err = s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, wsProject, wsName).Scan(&wsID)
	if err != nil {
		return fmt.Errorf("workstream not found: %s/%s", wsProject, wsName)
	}

	_, err = s.db.Exec(`INSERT INTO milestone_requirements (milestone_id, workstream_id) VALUES (?, ?)`, milestoneID, wsID)
	return err
}

// UpdateMilestoneDescription updates a milestone's description
func (s *Store) UpdateMilestoneDescription(project, name, description string) error {
	result, err := s.db.Exec(`UPDATE milestones SET description = ? WHERE project = ? AND name = ?`, description, project, name)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("milestone not found: %s/%s", project, name)
	}
	return nil
}

// RemoveMilestoneRequirement removes a workstream requirement from a milestone
func (s *Store) RemoveMilestoneRequirement(milestoneProject, milestoneName, wsProject, wsName string) error {
	var milestoneID, wsID int64

	err := s.db.QueryRow(`SELECT id FROM milestones WHERE project = ? AND name = ?`, milestoneProject, milestoneName).Scan(&milestoneID)
	if err != nil {
		return fmt.Errorf("milestone not found: %s/%s", milestoneProject, milestoneName)
	}

	err = s.db.QueryRow(`SELECT id FROM workstreams WHERE project = ? AND name = ?`, wsProject, wsName).Scan(&wsID)
	if err != nil {
		return fmt.Errorf("workstream not found: %s/%s", wsProject, wsName)
	}

	_, err = s.db.Exec(`DELETE FROM milestone_requirements WHERE milestone_id = ? AND workstream_id = ?`, milestoneID, wsID)
	return err
}

// DeleteMilestone removes a milestone. This does NOT delete associated workstreams -
// milestones are just groupings that reference workstreams, not owners of them.
func (s *Store) DeleteMilestone(project, name string) error {
	result, err := s.db.Exec(`DELETE FROM milestones WHERE project = ? AND name = ?`, project, name)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("milestone not found: %s/%s", project, name)
	}
	// Note: milestone_requirements are deleted via ON DELETE CASCADE
	return nil
}

// computeMilestoneStatus determines milestone status from requirements
func computeMilestoneStatus(reqs []workstream.MilestoneRequirement) workstream.State {
	if len(reqs) == 0 {
		return workstream.StatePending
	}

	doneCount := 0
	for _, req := range reqs {
		if req.WorkstreamState == workstream.StateDone {
			doneCount++
		}
	}

	if doneCount == len(reqs) {
		return workstream.StateDone
	}
	if doneCount > 0 {
		return workstream.StateInProgress
	}
	return workstream.StatePending
}

// MigrateLogNewlines fixes existing log entries that have escaped newlines
func (s *Store) MigrateLogNewlines() (int64, error) {
	// Fix \n escapes
	result1, err := s.db.Exec(`
		UPDATE log_entries
		SET content = REPLACE(content, '\n', char(10))
		WHERE content LIKE '%\n%'`)
	if err != nil {
		return 0, err
	}
	affected1, _ := result1.RowsAffected()

	// Fix \u000A escapes
	result2, err := s.db.Exec(`
		UPDATE log_entries
		SET content = REPLACE(content, '\u000A', char(10))
		WHERE content LIKE '%\u000A%'`)
	if err != nil {
		return affected1, err
	}
	affected2, _ := result2.RowsAffected()

	return affected1 + affected2, nil
}
