package store

import (
	"path/filepath"
	"testing"
	"time"

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
	if got.Log[0].Content != "First entry" {
		t.Errorf("Log[0].Content = %q, want 'First entry'", got.Log[0].Content)
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
