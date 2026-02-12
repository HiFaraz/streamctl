package store

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/faraz/streamctl/pkg/workstream"
)

func TestNewStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer s.Close()

	if s == nil {
		t.Fatal("New() returned nil store")
	}
}

func TestCreateAndGet(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	ws := &workstream.Workstream{
		Name:       "Test Feature",
		Project:    "myproject",
		State:      workstream.StatePending,
		Objective:  "Build something great.",
		LastUpdate: time.Now().UTC().Truncate(time.Second),
	}

	err := s.Create(ws)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := s.Get("myproject", "Test Feature")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Name != ws.Name {
		t.Errorf("Name = %q, want %q", got.Name, ws.Name)
	}
	if got.Project != ws.Project {
		t.Errorf("Project = %q, want %q", got.Project, ws.Project)
	}
	if got.Objective != ws.Objective {
		t.Errorf("Objective = %q, want %q", got.Objective, ws.Objective)
	}
}

func TestCreateWithPlanItems(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	ws := &workstream.Workstream{
		Name:      "Feature with Plan",
		Project:   "myproject",
		State:     workstream.StatePending,
		Objective: "Test plan items.",
		Plan: []workstream.PlanItem{
			{Text: "Step one", Complete: false},
			{Text: "Step two", Complete: true},
			{Text: "Step three", Complete: false},
		},
	}

	s.Create(ws)

	got, _ := s.Get("myproject", "Feature with Plan")
	if len(got.Plan) != 3 {
		t.Fatalf("Plan length = %d, want 3", len(got.Plan))
	}
	if got.Plan[0].Text != "Step one" || got.Plan[0].Complete {
		t.Errorf("Plan[0] = %+v, want Step one/false", got.Plan[0])
	}
	if got.Plan[1].Text != "Step two" || !got.Plan[1].Complete {
		t.Errorf("Plan[1] = %+v, want Step two/true", got.Plan[1])
	}
}

func TestCreateWithLogEntries(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	now := time.Now().UTC().Truncate(time.Second)
	ws := &workstream.Workstream{
		Name:      "Feature with Log",
		Project:   "myproject",
		State:     workstream.StatePending,
		Objective: "Test log entries.",
		Log: []workstream.LogEntry{
			{Timestamp: now.Add(-time.Hour), Content: "First entry"},
			{Timestamp: now, Content: "Second entry"},
		},
	}

	s.Create(ws)

	got, _ := s.Get("myproject", "Feature with Log")
	if len(got.Log) != 2 {
		t.Fatalf("Log length = %d, want 2", len(got.Log))
	}
	// Logs are returned newest first
	if got.Log[0].Content != "Second entry" {
		t.Errorf("Log[0].Content = %q, want 'Second entry' (newest first)", got.Log[0].Content)
	}
}

func TestListProjects(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	// Create workstreams in different projects
	s.Create(&workstream.Workstream{Name: "WS1", Project: "project-a", State: workstream.StatePending})
	s.Create(&workstream.Workstream{Name: "WS2", Project: "project-b", State: workstream.StatePending})
	s.Create(&workstream.Workstream{Name: "WS3", Project: "project-a", State: workstream.StatePending})

	projects, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("ListProjects() = %d projects, want 2", len(projects))
	}
}

func TestList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{Name: "WS1", Project: "proj", State: workstream.StatePending})
	s.Create(&workstream.Workstream{Name: "WS2", Project: "proj", State: workstream.StateInProgress})
	s.Create(&workstream.Workstream{Name: "WS3", Project: "other", State: workstream.StatePending})

	// List all
	all, _ := s.List(Filter{})
	if len(all) != 3 {
		t.Errorf("List() all = %d, want 3", len(all))
	}

	// Filter by project
	projOnly, _ := s.List(Filter{Project: "proj"})
	if len(projOnly) != 2 {
		t.Errorf("List(project=proj) = %d, want 2", len(projOnly))
	}

	// Filter by state
	inProgress, _ := s.List(Filter{State: workstream.StateInProgress})
	if len(inProgress) != 1 {
		t.Errorf("List(state=in_progress) = %d, want 1", len(inProgress))
	}
}

func TestListFilterByOwner(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{Name: "WS1", Project: "proj", State: workstream.StatePending, Owner: "agent-1"})
	s.Create(&workstream.Workstream{Name: "WS2", Project: "proj", State: workstream.StatePending, Owner: "agent-2"})
	s.Create(&workstream.Workstream{Name: "WS3", Project: "proj", State: workstream.StatePending})

	owned, _ := s.List(Filter{Owner: "agent-1"})
	if len(owned) != 1 {
		t.Errorf("List(owner=agent-1) = %d, want 1", len(owned))
	}
}

func TestUpdate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{
		Name:      "Update Test",
		Project:   "proj",
		State:     workstream.StatePending,
		Objective: "Original objective.",
	})

	// Update state
	newState := workstream.StateInProgress
	err := s.Update("proj", "Update Test", WorkstreamUpdate{State: &newState})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, _ := s.Get("proj", "Update Test")
	if got.State != workstream.StateInProgress {
		t.Errorf("State = %q, want in_progress", got.State)
	}
}

func TestUpdateOwner(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{Name: "Owner Test", Project: "proj", State: workstream.StatePending})

	owner := "agent-123"
	s.Update("proj", "Owner Test", WorkstreamUpdate{Owner: &owner})

	got, _ := s.Get("proj", "Owner Test")
	if got.Owner != "agent-123" {
		t.Errorf("Owner = %q, want agent-123", got.Owner)
	}

	// Clear owner
	emptyOwner := ""
	s.Update("proj", "Owner Test", WorkstreamUpdate{Owner: &emptyOwner})

	got, _ = s.Get("proj", "Owner Test")
	if got.Owner != "" {
		t.Errorf("Owner = %q, want empty", got.Owner)
	}
}

func TestUpdateLogEntry(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{Name: "Log Test", Project: "proj", State: workstream.StatePending})

	logEntry := "New log entry content."
	s.Update("proj", "Log Test", WorkstreamUpdate{LogEntry: &logEntry})

	got, _ := s.Get("proj", "Log Test")
	if len(got.Log) != 1 {
		t.Fatalf("Log length = %d, want 1", len(got.Log))
	}
	if got.Log[0].Content != logEntry {
		t.Errorf("Log[0].Content = %q, want %q", got.Log[0].Content, logEntry)
	}
}

func TestUpdatePlanItem(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{
		Name:    "Plan Test",
		Project: "proj",
		State:   workstream.StatePending,
		Plan: []workstream.PlanItem{
			{Text: "Step one", Complete: false},
			{Text: "Step two", Complete: false},
		},
	})

	// Toggle first item
	idx := 0
	s.Update("proj", "Plan Test", WorkstreamUpdate{PlanIndex: &idx})

	got, _ := s.Get("proj", "Plan Test")
	if !got.Plan[0].Complete {
		t.Errorf("Plan[0].Complete = false, want true")
	}

	// Toggle again
	s.Update("proj", "Plan Test", WorkstreamUpdate{PlanIndex: &idx})

	got, _ = s.Get("proj", "Plan Test")
	if got.Plan[0].Complete {
		t.Errorf("Plan[0].Complete = true, want false")
	}
}

func TestRename(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{Name: "Old Name", Project: "proj", State: workstream.StatePending, Objective: "Test objective"})

	err := s.Rename("proj", "Old Name", "New Name")
	if err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	// Should not find old name
	_, err = s.Get("proj", "Old Name")
	if err == nil {
		t.Errorf("Get() old name should return error after rename")
	}

	// Should find new name
	ws, err := s.Get("proj", "New Name")
	if err != nil {
		t.Fatalf("Get() new name error = %v", err)
	}
	if ws.Objective != "Test objective" {
		t.Errorf("Objective = %q, want %q", ws.Objective, "Test objective")
	}
}

func TestDelete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{Name: "Delete Test", Project: "proj", State: workstream.StatePending})

	err := s.Delete("proj", "Delete Test")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = s.Get("proj", "Delete Test")
	if err == nil {
		t.Errorf("Get() after Delete() should return error")
	}
}

func TestTaskNotes(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{
		Name:    "Notes Test",
		Project: "proj",
		State:   workstream.StatePending,
	})

	// Add task
	s.AddTask("proj", "Notes Test", "Task with notes")

	// Set notes with markdown
	notes := "## Details\n- Item 1\n- Item 2\n\n```go\nfunc main() {}\n```"
	err := s.SetTaskNotes("proj", "Notes Test", 0, notes)
	if err != nil {
		t.Fatalf("SetTaskNotes() error = %v", err)
	}

	got, _ := s.Get("proj", "Notes Test")
	if got.Plan[0].Notes != notes {
		t.Errorf("Notes = %q, want %q", got.Plan[0].Notes, notes)
	}
}

func TestSetTaskStatus(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{
		Name:    "Status Test",
		Project: "proj",
		State:   workstream.StatePending,
		Plan: []workstream.PlanItem{
			{Text: "Task 0", Status: workstream.TaskPending},
		},
	})

	// Set to in_progress
	err := s.SetTaskStatus("proj", "Status Test", 0, workstream.TaskInProgress)
	if err != nil {
		t.Fatalf("SetTaskStatus() error = %v", err)
	}

	got, _ := s.Get("proj", "Status Test")
	if got.Plan[0].Status != workstream.TaskInProgress {
		t.Errorf("Status = %q, want in_progress", got.Plan[0].Status)
	}

	// Set to done
	s.SetTaskStatus("proj", "Status Test", 0, workstream.TaskDone)
	got, _ = s.Get("proj", "Status Test")
	if got.Plan[0].Status != workstream.TaskDone {
		t.Errorf("Status = %q, want done", got.Plan[0].Status)
	}
	// Complete should also be true when done
	if !got.Plan[0].Complete {
		t.Errorf("Complete = false, want true when status is done")
	}
}

func TestRemoveTask(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{
		Name:    "Remove Test",
		Project: "proj",
		State:   workstream.StatePending,
		Plan: []workstream.PlanItem{
			{Text: "Task 0", Status: workstream.TaskPending},
			{Text: "Task 1", Status: workstream.TaskPending},
			{Text: "Task 2", Status: workstream.TaskPending},
		},
	})

	// Remove middle task
	err := s.RemoveTask("proj", "Remove Test", 1)
	if err != nil {
		t.Fatalf("RemoveTask() error = %v", err)
	}

	got, _ := s.Get("proj", "Remove Test")
	if len(got.Plan) != 2 {
		t.Fatalf("Plan length = %d, want 2", len(got.Plan))
	}
	if got.Plan[0].Text != "Task 0" {
		t.Errorf("Plan[0].Text = %q, want 'Task 0'", got.Plan[0].Text)
	}
	if got.Plan[1].Text != "Task 2" {
		t.Errorf("Plan[1].Text = %q, want 'Task 2'", got.Plan[1].Text)
	}
}

func TestAddTask(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{
		Name:    "Task Test",
		Project: "proj",
		State:   workstream.StatePending,
	})

	err := s.AddTask("proj", "Task Test", "New task")
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	got, _ := s.Get("proj", "Task Test")
	if len(got.Plan) != 1 {
		t.Fatalf("Plan length = %d, want 1", len(got.Plan))
	}
	if got.Plan[0].Text != "New task" {
		t.Errorf("Plan[0].Text = %q, want 'New task'", got.Plan[0].Text)
	}
	if got.Plan[0].Status != workstream.TaskPending {
		t.Errorf("Plan[0].Status = %q, want pending", got.Plan[0].Status)
	}

	// Add another task
	s.AddTask("proj", "Task Test", "Second task")
	got, _ = s.Get("proj", "Task Test")
	if len(got.Plan) != 2 {
		t.Fatalf("Plan length = %d, want 2", len(got.Plan))
	}
	if got.Plan[1].Text != "Second task" {
		t.Errorf("Plan[1].Text = %q, want 'Second task'", got.Plan[1].Text)
	}
}

func TestRemoveDependency(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{Name: "auth", Project: "proj", State: workstream.StatePending})
	s.Create(&workstream.Workstream{Name: "api", Project: "proj", State: workstream.StatePending})
	s.AddDependency("proj", "auth", "proj", "api")

	// Remove dependency
	err := s.RemoveDependency("proj", "auth", "proj", "api")
	if err != nil {
		t.Fatalf("RemoveDependency() error = %v", err)
	}

	// Verify api is no longer blocked
	api, _ := s.Get("proj", "api")
	if len(api.BlockedBy) != 0 {
		t.Errorf("BlockedBy length = %d, want 0", len(api.BlockedBy))
	}
}

func TestAddDependency(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	// Create two workstreams
	s.Create(&workstream.Workstream{Name: "auth", Project: "proj", State: workstream.StatePending})
	s.Create(&workstream.Workstream{Name: "api", Project: "proj", State: workstream.StatePending})

	// auth blocks api
	err := s.AddDependency("proj", "auth", "proj", "api")
	if err != nil {
		t.Fatalf("AddDependency() error = %v", err)
	}

	// Verify api is blocked by auth
	api, _ := s.Get("proj", "api")
	if len(api.BlockedBy) != 1 {
		t.Fatalf("BlockedBy length = %d, want 1", len(api.BlockedBy))
	}
	if api.BlockedBy[0].BlockerName != "auth" {
		t.Errorf("BlockedBy[0].BlockerName = %q, want 'auth'", api.BlockedBy[0].BlockerName)
	}

	// Verify auth blocks api
	auth, _ := s.Get("proj", "auth")
	if len(auth.Blocks) != 1 {
		t.Fatalf("Blocks length = %d, want 1", len(auth.Blocks))
	}
	if auth.Blocks[0].BlockedName != "api" {
		t.Errorf("Blocks[0].BlockedName = %q, want 'api'", auth.Blocks[0].BlockedName)
	}
}

func TestMigrateExistingDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Create old-style database manually (without status or notes columns)
	db, _ := sql.Open("sqlite3", dbPath)
	db.Exec(`
		CREATE TABLE workstreams (
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
		CREATE TABLE plan_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workstream_id INTEGER NOT NULL,
			position INTEGER NOT NULL,
			text TEXT NOT NULL,
			complete BOOLEAN DEFAULT FALSE
		);
		CREATE TABLE log_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workstream_id INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			content TEXT NOT NULL
		);
	`)
	// Insert old-style data
	db.Exec(`INSERT INTO workstreams (project, name, state, last_update) VALUES ('proj', 'test', 'pending', datetime('now'))`)
	db.Exec(`INSERT INTO plan_items (workstream_id, position, text, complete) VALUES (1, 0, 'Old task', 1)`)
	db.Close()

	// Open with store (should migrate)
	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer s.Close()

	// Verify status column exists and old data migrated
	ws, err := s.Get("proj", "test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if len(ws.Plan) != 1 {
		t.Fatalf("Plan length = %d, want 1", len(ws.Plan))
	}
	// Old complete=true should become status=done
	if ws.Plan[0].Status != workstream.TaskDone {
		t.Errorf("Status = %q, want 'done' (migrated from complete=true)", ws.Plan[0].Status)
	}
	// Notes should default to empty
	if ws.Plan[0].Notes != "" {
		t.Errorf("Notes = %q, want empty string", ws.Plan[0].Notes)
	}

	// Verify dependencies table exists
	_, err = s.db.Exec("SELECT 1 FROM workstream_dependencies LIMIT 1")
	if err != nil {
		t.Errorf("workstream_dependencies table missing after migration: %v", err)
	}
}

func TestDependenciesTableExists(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	// The table should exist after New() runs migrations
	_, err := s.db.Exec("SELECT 1 FROM workstream_dependencies LIMIT 1")
	if err != nil {
		t.Errorf("workstream_dependencies table does not exist: %v", err)
	}
}

func TestPlanItemStatus(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	ws := &workstream.Workstream{
		Name:    "Status Test",
		Project: "proj",
		State:   workstream.StatePending,
		Plan: []workstream.PlanItem{
			{Text: "Task 1", Status: workstream.TaskPending},
			{Text: "Task 2", Status: workstream.TaskInProgress},
			{Text: "Task 3", Status: workstream.TaskDone},
			{Text: "Task 4", Status: workstream.TaskSkipped},
		},
	}

	err := s.Create(ws)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := s.Get("proj", "Status Test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if len(got.Plan) != 4 {
		t.Fatalf("Plan length = %d, want 4", len(got.Plan))
	}
	if got.Plan[0].Status != workstream.TaskPending {
		t.Errorf("Plan[0].Status = %q, want pending", got.Plan[0].Status)
	}
	if got.Plan[1].Status != workstream.TaskInProgress {
		t.Errorf("Plan[1].Status = %q, want in_progress", got.Plan[1].Status)
	}
	if got.Plan[2].Status != workstream.TaskDone {
		t.Errorf("Plan[2].Status = %q, want done", got.Plan[2].Status)
	}
	if got.Plan[3].Status != workstream.TaskSkipped {
		t.Errorf("Plan[3].Status = %q, want skipped", got.Plan[3].Status)
	}
}

func TestRecentActivity(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	// Create workstreams with log entries
	s.Create(&workstream.Workstream{
		Name: "ws1", Project: "proj", State: workstream.StatePending,
		Log: []workstream.LogEntry{
			{Timestamp: time.Now().Add(-2 * time.Hour), Content: "Old entry"},
		},
	})
	s.Create(&workstream.Workstream{
		Name: "ws2", Project: "proj", State: workstream.StatePending,
		Log: []workstream.LogEntry{
			{Timestamp: time.Now().Add(-1 * time.Hour), Content: "Recent entry"},
		},
	})

	// Add another entry
	logEntry := "Newest entry"
	s.Update("proj", "ws1", WorkstreamUpdate{LogEntry: &logEntry})

	activity, err := s.RecentActivity("proj", 10, 0)
	if err != nil {
		t.Fatalf("RecentActivity() error = %v", err)
	}

	if len(activity) != 3 {
		t.Fatalf("RecentActivity() = %d entries, want 3", len(activity))
	}

	// Should be ordered by timestamp descending (newest first)
	if activity[0].Content != "Newest entry" {
		t.Errorf("activity[0].Content = %q, want 'Newest entry'", activity[0].Content)
	}
}

func TestNeedsHelp(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{
		Name: "test", Project: "proj", State: workstream.StateInProgress,
	})

	// Initially false
	ws, _ := s.Get("proj", "test")
	if ws.NeedsHelp {
		t.Error("NeedsHelp should be false initially")
	}

	// Set to true
	needsHelp := true
	s.Update("proj", "test", WorkstreamUpdate{NeedsHelp: &needsHelp})

	ws, _ = s.Get("proj", "test")
	if !ws.NeedsHelp {
		t.Error("NeedsHelp should be true after update")
	}

	// Set back to false
	needsHelp = false
	s.Update("proj", "test", WorkstreamUpdate{NeedsHelp: &needsHelp})

	ws, _ = s.Get("proj", "test")
	if ws.NeedsHelp {
		t.Error("NeedsHelp should be false after clearing")
	}
}

func TestCreateAndGetMilestone(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	m := &workstream.Milestone{
		Name:        "wave-1-complete",
		Project:     "myproject",
		Description: "Foundation layer complete",
	}

	err := s.CreateMilestone(m)
	if err != nil {
		t.Fatalf("CreateMilestone() error = %v", err)
	}

	got, err := s.GetMilestone("myproject", "wave-1-complete")
	if err != nil {
		t.Fatalf("GetMilestone() error = %v", err)
	}

	if got.Name != m.Name {
		t.Errorf("Name = %q, want %q", got.Name, m.Name)
	}
	if got.Project != m.Project {
		t.Errorf("Project = %q, want %q", got.Project, m.Project)
	}
	if got.Description != m.Description {
		t.Errorf("Description = %q, want %q", got.Description, m.Description)
	}
	// Status should be pending (no requirements)
	if got.Status != workstream.StatePending {
		t.Errorf("Status = %q, want pending", got.Status)
	}
}

func TestMilestoneRequirements(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	// Create workstreams
	s.Create(&workstream.Workstream{Name: "auth", Project: "proj", State: workstream.StatePending})
	s.Create(&workstream.Workstream{Name: "api", Project: "proj", State: workstream.StatePending})

	// Create milestone
	s.CreateMilestone(&workstream.Milestone{Name: "wave-1", Project: "proj"})

	// Add requirements
	err := s.AddMilestoneRequirement("proj", "wave-1", "proj", "auth")
	if err != nil {
		t.Fatalf("AddMilestoneRequirement() error = %v", err)
	}
	err = s.AddMilestoneRequirement("proj", "wave-1", "proj", "api")
	if err != nil {
		t.Fatalf("AddMilestoneRequirement() error = %v", err)
	}

	// Verify requirements
	m, _ := s.GetMilestone("proj", "wave-1")
	if len(m.Requirements) != 2 {
		t.Fatalf("Requirements length = %d, want 2", len(m.Requirements))
	}
	if m.Status != workstream.StatePending {
		t.Errorf("Status = %q, want pending (no workstreams done)", m.Status)
	}
}

func TestMilestoneStatusComputation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	// Create workstreams
	s.Create(&workstream.Workstream{Name: "ws-a", Project: "proj", State: workstream.StatePending})
	s.Create(&workstream.Workstream{Name: "ws-b", Project: "proj", State: workstream.StatePending})

	// Create milestone with requirements
	s.CreateMilestone(&workstream.Milestone{Name: "gate", Project: "proj"})
	s.AddMilestoneRequirement("proj", "gate", "proj", "ws-a")
	s.AddMilestoneRequirement("proj", "gate", "proj", "ws-b")

	// Status should be pending (all pending)
	m, _ := s.GetMilestone("proj", "gate")
	if m.Status != workstream.StatePending {
		t.Errorf("Status = %q, want pending", m.Status)
	}

	// Complete one workstream -> in_progress
	doneState := workstream.StateDone
	s.Update("proj", "ws-a", WorkstreamUpdate{State: &doneState})

	m, _ = s.GetMilestone("proj", "gate")
	if m.Status != workstream.StateInProgress {
		t.Errorf("Status = %q, want in_progress", m.Status)
	}

	// Complete second workstream -> done
	s.Update("proj", "ws-b", WorkstreamUpdate{State: &doneState})

	m, _ = s.GetMilestone("proj", "gate")
	if m.Status != workstream.StateDone {
		t.Errorf("Status = %q, want done", m.Status)
	}
}

func TestListMilestones(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	// Create workstreams and milestones
	s.Create(&workstream.Workstream{Name: "ws-1", Project: "proj", State: workstream.StateDone})
	s.Create(&workstream.Workstream{Name: "ws-2", Project: "proj", State: workstream.StatePending})
	s.Create(&workstream.Workstream{Name: "ws-3", Project: "other", State: workstream.StatePending})

	s.CreateMilestone(&workstream.Milestone{Name: "gate-1", Project: "proj"})
	s.CreateMilestone(&workstream.Milestone{Name: "gate-2", Project: "proj"})
	s.CreateMilestone(&workstream.Milestone{Name: "gate-3", Project: "other"})

	s.AddMilestoneRequirement("proj", "gate-1", "proj", "ws-1")
	s.AddMilestoneRequirement("proj", "gate-2", "proj", "ws-2")

	// List all
	all, err := s.ListMilestones("")
	if err != nil {
		t.Fatalf("ListMilestones() error = %v", err)
	}
	if len(all) != 3 {
		t.Errorf("ListMilestones() = %d, want 3", len(all))
	}

	// Filter by project
	projOnly, _ := s.ListMilestones("proj")
	if len(projOnly) != 2 {
		t.Errorf("ListMilestones(proj) = %d, want 2", len(projOnly))
	}

	// gate-1 should be done (ws-1 is done)
	for _, m := range projOnly {
		if m.Name == "gate-1" && m.Status != workstream.StateDone {
			t.Errorf("gate-1 status = %q, want done", m.Status)
		}
		if m.Name == "gate-2" && m.Status != workstream.StatePending {
			t.Errorf("gate-2 status = %q, want pending", m.Status)
		}
	}
}

func TestRemoveMilestoneRequirement(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.Create(&workstream.Workstream{Name: "auth", Project: "proj", State: workstream.StatePending})
	s.Create(&workstream.Workstream{Name: "api", Project: "proj", State: workstream.StatePending})

	s.CreateMilestone(&workstream.Milestone{Name: "gate", Project: "proj"})
	s.AddMilestoneRequirement("proj", "gate", "proj", "auth")
	s.AddMilestoneRequirement("proj", "gate", "proj", "api")

	// Verify 2 requirements
	m, _ := s.GetMilestone("proj", "gate")
	if len(m.Requirements) != 2 {
		t.Fatalf("Requirements = %d, want 2", len(m.Requirements))
	}

	// Remove one
	err := s.RemoveMilestoneRequirement("proj", "gate", "proj", "auth")
	if err != nil {
		t.Fatalf("RemoveMilestoneRequirement() error = %v", err)
	}

	m, _ = s.GetMilestone("proj", "gate")
	if len(m.Requirements) != 1 {
		t.Fatalf("Requirements = %d, want 1 after removal", len(m.Requirements))
	}
	if m.Requirements[0].WorkstreamName != "api" {
		t.Errorf("Remaining requirement = %q, want 'api'", m.Requirements[0].WorkstreamName)
	}
}

func TestUpdateMilestoneDescription(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, _ := New(dbPath)
	defer s.Close()

	s.CreateMilestone(&workstream.Milestone{Name: "gate", Project: "proj", Description: "Original"})

	err := s.UpdateMilestoneDescription("proj", "gate", "Updated description")
	if err != nil {
		t.Fatalf("UpdateMilestoneDescription() error = %v", err)
	}

	m, _ := s.GetMilestone("proj", "gate")
	if m.Description != "Updated description" {
		t.Errorf("Description = %q, want 'Updated description'", m.Description)
	}
}
