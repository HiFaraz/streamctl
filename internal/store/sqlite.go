package store

import (
	"database/sql"
	"fmt"
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

// migrate creates the database schema
func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS workstreams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project TEXT NOT NULL,
		name TEXT NOT NULL,
		state TEXT NOT NULL DEFAULT 'pending',
		owner TEXT DEFAULT '',
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
		complete BOOLEAN DEFAULT FALSE
	);

	CREATE TABLE IF NOT EXISTS log_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workstream_id INTEGER NOT NULL REFERENCES workstreams(id) ON DELETE CASCADE,
		timestamp DATETIME NOT NULL,
		content TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_workstreams_project ON workstreams(project);
	CREATE INDEX IF NOT EXISTS idx_workstreams_state ON workstreams(state);
	CREATE INDEX IF NOT EXISTS idx_workstreams_owner ON workstreams(owner);
	CREATE INDEX IF NOT EXISTS idx_plan_items_workstream ON plan_items(workstream_id);
	CREATE INDEX IF NOT EXISTS idx_log_entries_workstream ON log_entries(workstream_id);
	`
	_, err := s.db.Exec(schema)
	return err
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
		INSERT INTO workstreams (project, name, state, owner, objective, key_context, decisions, last_update)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		ws.Project, ws.Name, string(ws.State), ws.Owner, ws.Objective, ws.KeyContext, ws.Decisions, ws.LastUpdate,
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
		_, err := tx.Exec(`
			INSERT INTO plan_items (workstream_id, position, text, complete)
			VALUES (?, ?, ?, ?)`,
			wsID, i, item.Text, item.Complete,
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
		SELECT id, project, name, state, owner, objective, key_context, decisions, last_update
		FROM workstreams WHERE project = ? AND name = ?`,
		project, name,
	).Scan(&wsID, &ws.Project, &ws.Name, &ws.State, &ws.Owner, &ws.Objective, &ws.KeyContext, &ws.Decisions, &ws.LastUpdate)
	if err != nil {
		return nil, err
	}

	// Load plan items
	rows, err := s.db.Query(`
		SELECT text, complete FROM plan_items
		WHERE workstream_id = ? ORDER BY position`,
		wsID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item workstream.PlanItem
		if err := rows.Scan(&item.Text, &item.Complete); err != nil {
			return nil, err
		}
		ws.Plan = append(ws.Plan, item)
	}

	// Load log entries
	logRows, err := s.db.Query(`
		SELECT timestamp, content FROM log_entries
		WHERE workstream_id = ? ORDER BY timestamp`,
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
	query := `SELECT id, project, name, state, owner, objective, key_context, decisions, last_update FROM workstreams WHERE 1=1`
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
		if err := rows.Scan(&wsID, &ws.Project, &ws.Name, &ws.State, &ws.Owner, &ws.Objective, &ws.KeyContext, &ws.Decisions, &ws.LastUpdate); err != nil {
			return nil, err
		}

		// Load plan items
		planRows, err := s.db.Query(`SELECT text, complete FROM plan_items WHERE workstream_id = ? ORDER BY position`, wsID)
		if err != nil {
			return nil, err
		}
		for planRows.Next() {
			var item workstream.PlanItem
			planRows.Scan(&item.Text, &item.Complete)
			ws.Plan = append(ws.Plan, item)
		}
		planRows.Close()

		// Load log entries
		logRows, err := s.db.Query(`SELECT timestamp, content FROM log_entries WHERE workstream_id = ? ORDER BY timestamp`, wsID)
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

	// Append log entry
	if updates.LogEntry != nil {
		_, err := tx.Exec(`INSERT INTO log_entries (workstream_id, timestamp, content) VALUES (?, ?, ?)`,
			wsID, time.Now().UTC(), *updates.LogEntry)
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
