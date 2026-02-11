package web

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faraz/streamctl/internal/store"
	"github.com/faraz/streamctl/pkg/workstream"
)

func setupTestStore(t *testing.T) *store.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestServer_Index_ShowsProjectName(t *testing.T) {
	st := setupTestStore(t)
	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "myproject") {
		t.Errorf("body should contain project name 'myproject'")
	}
}

func TestServer_Index_ListsWorkstreams(t *testing.T) {
	st := setupTestStore(t)

	// Create a workstream
	ws := &workstream.Workstream{
		Project:   "myproject",
		Name:      "auth",
		State:     workstream.StateInProgress,
		Objective: "Add authentication",
	}
	if err := st.Create(ws); err != nil {
		t.Fatalf("Create: %v", err)
	}

	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "auth") {
		t.Errorf("body should contain workstream name 'auth', got:\n%s", body)
	}
	if !strings.Contains(body, "in_progress") {
		t.Errorf("body should contain state 'in_progress'")
	}
}

func TestServer_Workstream_ShowsDetails(t *testing.T) {
	st := setupTestStore(t)

	ws := &workstream.Workstream{
		Project:   "myproject",
		Name:      "auth",
		State:     workstream.StateInProgress,
		Objective: "Add authentication",
		Plan: []workstream.PlanItem{
			{Text: "Design schema", Status: workstream.TaskDone},
			{Text: "Implement login", Status: workstream.TaskPending},
		},
	}
	if err := st.Create(ws); err != nil {
		t.Fatalf("Create: %v", err)
	}

	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/workstream/auth", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Add authentication") {
		t.Errorf("body should contain objective")
	}
	if !strings.Contains(body, "Design schema") {
		t.Errorf("body should contain plan item")
	}
}
