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

func TestServer_Index_HasArrowRightNavigation(t *testing.T) {
	st := setupTestStore(t)
	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	// Right arrow should open selected item (same as Enter)
	if !strings.Contains(body, "ArrowRight") {
		t.Errorf("index page should handle ArrowRight key for navigation")
	}
}

func TestServer_Workstream_HasArrowLeftNavigation(t *testing.T) {
	st := setupTestStore(t)

	ws := &workstream.Workstream{
		Project: "myproject",
		Name:    "test-ws",
		State:   workstream.StatePending,
	}
	if err := st.Create(ws); err != nil {
		t.Fatalf("Create: %v", err)
	}

	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/workstream/test-ws", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	// Left arrow should go back (same as Backspace)
	if !strings.Contains(body, "ArrowLeft") {
		t.Errorf("workstream page should handle ArrowLeft key for back navigation")
	}
}

func TestServer_Workstream_HasMarkdownRendering(t *testing.T) {
	st := setupTestStore(t)

	ws := &workstream.Workstream{
		Project: "myproject",
		Name:    "md-test",
		State:   workstream.StatePending,
	}
	if err := st.Create(ws); err != nil {
		t.Fatalf("Create: %v", err)
	}

	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/workstream/md-test", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	// Should include marked.js for markdown rendering
	if !strings.Contains(body, "marked") {
		t.Errorf("workstream page should include marked.js for markdown rendering")
	}
}

func TestServer_Workstream_HasCollapsibleLogs(t *testing.T) {
	st := setupTestStore(t)

	ws := &workstream.Workstream{
		Project: "myproject",
		Name:    "collapse-test",
		State:   workstream.StatePending,
	}
	if err := st.Create(ws); err != nil {
		t.Fatalf("Create: %v", err)
	}

	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/workstream/collapse-test", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	// Should have expand/collapse functionality for logs
	if !strings.Contains(body, "expandLog") || !strings.Contains(body, "collapseLog") {
		t.Errorf("workstream page should have expandLog/collapseLog functions for collapsible logs")
	}
}

func TestServer_Workstream_KbdVisibleWhenSelected(t *testing.T) {
	st := setupTestStore(t)

	ws := &workstream.Workstream{
		Project: "myproject",
		Name:    "kbd-test",
		State:   workstream.StatePending,
	}
	if err := st.Create(ws); err != nil {
		t.Fatalf("Create: %v", err)
	}

	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/workstream/kbd-test", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	// kbd elements inside selected items should have proper contrast
	if !strings.Contains(body, ".feed-item.selected kbd") {
		t.Errorf("workstream page should style kbd elements inside selected items for visibility")
	}
}

func TestServer_Workstream_LogContentInScriptTag(t *testing.T) {
	st := setupTestStore(t)

	ws := &workstream.Workstream{
		Project: "myproject",
		Name:    "script-test",
		State:   workstream.StatePending,
	}
	if err := st.Create(ws); err != nil {
		t.Fatalf("Create: %v", err)
	}

	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/workstream/script-test", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	// Log content should be stored in a script tag to preserve newlines
	if !strings.Contains(body, "logContents") {
		t.Errorf("workstream page should store log contents in JavaScript to preserve newlines")
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

func TestServer_Workstream_LongObjective_IsFeedItem(t *testing.T) {
	st := setupTestStore(t)

	// Create a workstream with a long objective (> 200 chars)
	longObjective := "This is a very long objective that describes in detail what " +
		"the workstream aims to accomplish. It includes multiple paragraphs and " +
		"detailed information about the goals, requirements, and expected outcomes " +
		"of this particular piece of work. The objective is intentionally long to " +
		"test the expandable panel functionality."

	ws := &workstream.Workstream{
		Project:   "myproject",
		Name:      "long-obj",
		State:     workstream.StateInProgress,
		Objective: longObjective,
	}
	if err := st.Create(ws); err != nil {
		t.Fatalf("Create: %v", err)
	}

	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/workstream/long-obj", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()

	// Objective should be a feed item with data-type="objective"
	if !strings.Contains(body, `data-type="objective"`) {
		t.Errorf("long objective should be rendered as a feed item with data-type='objective'")
	}

	// Should have the objective pane element (not overlay)
	if !strings.Contains(body, "objective-pane") {
		t.Errorf("workstream page should have objective-pane element for split view")
	}

	// Should have badge-objective styling
	if !strings.Contains(body, "badge-objective") {
		t.Errorf("objective feed item should have badge-objective class")
	}
}

func TestServer_Workstream_ShortObjective_NotInFeed(t *testing.T) {
	st := setupTestStore(t)

	// Create a workstream with a short objective (< 200 chars)
	ws := &workstream.Workstream{
		Project:   "myproject",
		Name:      "short-obj",
		State:     workstream.StateInProgress,
		Objective: "A brief objective that fits on one line.",
	}
	if err := st.Create(ws); err != nil {
		t.Fatalf("Create: %v", err)
	}

	srv := NewServer(st, "myproject")

	req := httptest.NewRequest("GET", "/workstream/short-obj", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()

	// Should show the full objective in the header
	if !strings.Contains(body, "A brief objective that fits on one line.") {
		t.Errorf("body should contain full short objective")
	}

	// Short objectives should NOT be in the feed as a separate item
	// Check for the actual HTML article element, not JS selectors
	if strings.Contains(body, `<article class="feed-item`) && strings.Contains(body, `data-type="objective"`) {
		// More specific check - the actual feed item element
		if strings.Contains(body, `class="feed-item selected" data-index="0" data-type="objective"`) ||
			strings.Contains(body, `class="feed-item" data-index="0" data-type="objective"`) {
			t.Errorf("short objective should not have an objective feed item in the HTML")
		}
	}
}
