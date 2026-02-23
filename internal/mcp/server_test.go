package mcp

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/faraz/streamctl/internal/store"
	"github.com/faraz/streamctl/pkg/workstream"
	"github.com/mark3labs/mcp-go/mcp"
)

func setupTestStore(t *testing.T) *store.Store {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	// Create test workstreams
	s.Create(&workstream.Workstream{
		Name:       "Feature One",
		Project:    "testproject",
		State:      workstream.StatePending,
		Objective:  "First feature.",
		LastUpdate: time.Now(),
		Plan: []workstream.PlanItem{
			{Text: "Step one", Complete: false},
		},
	})

	s.Create(&workstream.Workstream{
		Name:       "Feature Two",
		Project:    "testproject",
		State:      workstream.StateInProgress,
		Owner:      "agent-123",
		Objective:  "Second feature.",
		LastUpdate: time.Now(),
		Plan: []workstream.PlanItem{
			{Text: "Done step", Complete: true},
		},
		Log: []workstream.LogEntry{
			{Timestamp: time.Now(), Content: "Started work."},
		},
	})

	return s
}

func TestHandleList(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{}
	result, err := h.HandleList(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleList() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleList() returned error result")
	}
}

func TestHandleListFilterProject(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
			},
		},
	}
	result, err := h.HandleList(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleList() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleList() returned error result")
	}
}

func TestHandleListFilterState(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"state": "in_progress",
			},
		},
	}
	result, err := h.HandleList(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleList() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleList() returned error result")
	}
}

func TestHandleGet(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
				"name":    "Feature One",
			},
		},
	}
	result, err := h.HandleGet(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleGet() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleGet() returned error result")
	}
}

func TestHandleCreate(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project":   "testproject",
				"name":      "new-feature",
				"objective": "Build a new thing.",
			},
		},
	}
	result, err := h.HandleCreate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleCreate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleCreate() returned error result")
	}

	// Verify it was created
	ws, err := st.Get("testproject", "new-feature")
	if err != nil {
		t.Fatalf("Get() after Create error = %v", err)
	}
	if ws.Objective != "Build a new thing." {
		t.Errorf("Objective = %q, want %q", ws.Objective, "Build a new thing.")
	}
}

func TestHandleUpdate(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
				"name":    "Feature One",
				"state":   "in_progress",
			},
		},
	}
	result, err := h.HandleUpdate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleUpdate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleUpdate() returned error result")
	}

	// Verify update
	ws, _ := st.Get("testproject", "Feature One")
	if ws.State != workstream.StateInProgress {
		t.Errorf("State = %q, want %q", ws.State, workstream.StateInProgress)
	}
}

func TestHandleClaim(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
				"name":    "Feature One",
				"owner":   "agent-456",
			},
		},
	}
	result, err := h.HandleClaim(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleClaim() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleClaim() returned error result")
	}

	// Verify owner set
	ws, _ := st.Get("testproject", "Feature One")
	if ws.Owner != "agent-456" {
		t.Errorf("Owner = %q, want %q", ws.Owner, "agent-456")
	}
}

func TestHandleRelease(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
				"name":    "Feature Two", // Has owner agent-123
			},
		},
	}
	result, err := h.HandleRelease(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRelease() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleRelease() returned error result")
	}

	// Verify owner cleared
	ws, _ := st.Get("testproject", "Feature Two")
	if ws.Owner != "" {
		t.Errorf("Owner = %q, want empty", ws.Owner)
	}
}

func TestHandleUpdateTaskAdd(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project":  "testproject",
				"name":     "Feature One",
				"task_add": "New task added via MCP",
			},
		},
	}
	result, err := h.HandleUpdate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleUpdate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleUpdate() returned error result")
	}

	// Verify task added
	ws, _ := st.Get("testproject", "Feature One")
	if len(ws.Plan) != 2 {
		t.Fatalf("Plan length = %d, want 2", len(ws.Plan))
	}
	if ws.Plan[1].Text != "New task added via MCP" {
		t.Errorf("Plan[1].Text = %q, want 'New task added via MCP'", ws.Plan[1].Text)
	}
}

func TestHandleUpdateTaskRemove(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project":     "testproject",
				"name":        "Feature One",
				"task_remove": float64(0), // JSON numbers come as float64
			},
		},
	}
	result, err := h.HandleUpdate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleUpdate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleUpdate() returned error result")
	}

	// Verify task removed
	ws, _ := st.Get("testproject", "Feature One")
	if len(ws.Plan) != 0 {
		t.Fatalf("Plan length = %d, want 0", len(ws.Plan))
	}
}

func TestHandleUpdateTaskStatus(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
				"name":    "Feature One",
				"task_status": map[string]any{
					"position": float64(0),
					"status":   "in_progress",
				},
			},
		},
	}
	result, err := h.HandleUpdate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleUpdate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleUpdate() returned error result")
	}

	// Verify status changed
	ws, _ := st.Get("testproject", "Feature One")
	if ws.Plan[0].Status != workstream.TaskInProgress {
		t.Errorf("Status = %q, want in_progress", ws.Plan[0].Status)
	}
}

func TestHandleUpdateTaskNotes(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
				"name":    "Feature One",
				"task_notes": map[string]any{
					"position": float64(0),
					"notes":    "## Details\n- Item 1\n```go\ncode\n```",
				},
			},
		},
	}
	result, err := h.HandleUpdate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleUpdate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleUpdate() returned error result")
	}

	// Verify notes set
	ws, _ := st.Get("testproject", "Feature One")
	if ws.Plan[0].Notes != "## Details\n- Item 1\n```go\ncode\n```" {
		t.Errorf("Notes = %q, want markdown content", ws.Plan[0].Notes)
	}
}

func TestHandleUpdateAddBlocker(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	// Feature One blocks Feature Two
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project":     "testproject",
				"name":        "Feature Two",
				"add_blocker": "testproject/Feature One",
			},
		},
	}
	result, err := h.HandleUpdate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleUpdate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleUpdate() returned error result")
	}

	// Verify dependency added
	ws, _ := st.Get("testproject", "Feature Two")
	if len(ws.BlockedBy) != 1 {
		t.Fatalf("BlockedBy length = %d, want 1", len(ws.BlockedBy))
	}
	if ws.BlockedBy[0].BlockerName != "Feature One" {
		t.Errorf("BlockerName = %q, want 'Feature One'", ws.BlockedBy[0].BlockerName)
	}
}

func TestHandleUpdateRemoveBlocker(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	// First add dependency
	st.AddDependency("testproject", "Feature One", "testproject", "Feature Two")

	// Then remove it
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project":        "testproject",
				"name":           "Feature Two",
				"remove_blocker": "testproject/Feature One",
			},
		},
	}
	result, err := h.HandleUpdate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleUpdate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleUpdate() returned error result")
	}

	// Verify dependency removed
	ws, _ := st.Get("testproject", "Feature Two")
	if len(ws.BlockedBy) != 0 {
		t.Errorf("BlockedBy length = %d, want 0", len(ws.BlockedBy))
	}
}

func TestHandleWebServe(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
			},
		},
	}

	result, err := h.HandleWebServe(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleWebServe() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleWebServe() returned error result")
	}

	// Result should contain a URL with project.localhost format
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "http://testproject.localhost:") {
		t.Errorf("Result should contain project.localhost URL, got: %s", text)
	}
}

func TestHandleMilestoneCreate(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project":     "testproject",
				"name":        "wave-1",
				"description": "Foundation layer",
			},
		},
	}
	result, err := h.HandleMilestoneCreate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleMilestoneCreate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleMilestoneCreate() returned error result")
	}

	// Verify created
	m, err := st.GetMilestone("testproject", "wave-1")
	if err != nil {
		t.Fatalf("GetMilestone() error = %v", err)
	}
	if m.Description != "Foundation layer" {
		t.Errorf("Description = %q, want 'Foundation layer'", m.Description)
	}
}

func TestHandleMilestoneGet(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	// Create milestone with requirements
	st.CreateMilestone(&workstream.Milestone{Name: "gate", Project: "testproject"})
	st.AddMilestoneRequirement("testproject", "gate", "testproject", "Feature One")

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
				"name":    "gate",
			},
		},
	}
	result, err := h.HandleMilestoneGet(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleMilestoneGet() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleMilestoneGet() returned error result")
	}

	// Should contain milestone info
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "gate") {
		t.Errorf("Result should contain milestone name, got: %s", text)
	}
}

func TestHandleMilestoneList(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	st.CreateMilestone(&workstream.Milestone{Name: "gate-1", Project: "testproject"})
	st.CreateMilestone(&workstream.Milestone{Name: "gate-2", Project: "testproject"})

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
			},
		},
	}
	result, err := h.HandleMilestoneList(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleMilestoneList() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleMilestoneList() returned error result")
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "gate-1") || !strings.Contains(text, "gate-2") {
		t.Errorf("Result should contain both milestones, got: %s", text)
	}
}

func TestHandleMilestoneUpdate(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	st.CreateMilestone(&workstream.Milestone{Name: "gate", Project: "testproject"})

	// Add requirement
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project":         "testproject",
				"name":            "gate",
				"add_requirement": "testproject/Feature One",
			},
		},
	}
	result, err := h.HandleMilestoneUpdate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleMilestoneUpdate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleMilestoneUpdate() returned error result")
	}

	m, _ := st.GetMilestone("testproject", "gate")
	if len(m.Requirements) != 1 {
		t.Fatalf("Requirements = %d, want 1", len(m.Requirements))
	}

	// Update description
	req = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project":     "testproject",
				"name":        "gate",
				"description": "Updated desc",
			},
		},
	}
	result, err = h.HandleMilestoneUpdate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleMilestoneUpdate() error = %v", err)
	}

	m, _ = st.GetMilestone("testproject", "gate")
	if m.Description != "Updated desc" {
		t.Errorf("Description = %q, want 'Updated desc'", m.Description)
	}
}

func TestHandleMilestoneDelete(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	// Create milestone
	st.CreateMilestone(&workstream.Milestone{Name: "to-delete", Project: "testproject"})
	st.AddMilestoneRequirement("testproject", "to-delete", "testproject", "Feature One")

	// Delete it
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
				"name":    "to-delete",
			},
		},
	}
	result, err := h.HandleMilestoneDelete(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleMilestoneDelete() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleMilestoneDelete() returned error result")
	}

	// Milestone should be gone
	_, err = st.GetMilestone("testproject", "to-delete")
	if err == nil {
		t.Error("Milestone should have been deleted")
	}

	// Workstream should still exist
	ws, err := st.Get("testproject", "Feature One")
	if err != nil {
		t.Fatalf("Workstream was deleted when milestone was deleted: %v", err)
	}
	if ws.Name != "Feature One" {
		t.Errorf("ws.Name = %q, want 'Feature One'", ws.Name)
	}
}

func TestHandleMilestoneDelete_NotFound(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
				"name":    "nonexistent",
			},
		},
	}
	result, err := h.HandleMilestoneDelete(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleMilestoneDelete() error = %v", err)
	}
	if !result.IsError {
		t.Error("HandleMilestoneDelete() should return error for non-existent milestone")
	}
}

func TestHandleBugList(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	// Add bugs to the test workstreams
	st.AddBug("testproject", "Feature One", "Bug in feature one", "agent-1")
	st.AddBug("testproject", "Feature Two", "Bug in feature two", "agent-2")

	req := mcp.CallToolRequest{}
	result, err := h.HandleBugList(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleBugList() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleBugList() returned error result")
	}

	// Result should contain JSON with both bugs
	content := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(content, "Bug in feature one") {
		t.Error("HandleBugList() should return 'Bug in feature one'")
	}
	if !strings.Contains(content, "Bug in feature two") {
		t.Error("HandleBugList() should return 'Bug in feature two'")
	}
}

func TestHandleBugReport(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project":     "testproject",
				"workstream":  "Feature One",
				"description": "Auth token not refreshed",
				"reported_by": "review-agent",
			},
		},
	}
	result, err := h.HandleBugReport(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleBugReport() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleBugReport() returned error result")
	}

	// Verify bug was added
	ws, _ := st.Get("testproject", "Feature One")
	found := false
	for _, item := range ws.Plan {
		if item.IsBug && item.Text == "Auth token not refreshed" && item.ReportedBy == "review-agent" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Bug was not added to workstream")
	}
}

func TestHandleListTruncatesLongObjectives(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	// Create workstream with very long objective (500+ chars)
	longObjective := strings.Repeat("This is a long objective. ", 50) // ~1250 chars
	st.Create(&workstream.Workstream{
		Name:       "long-objective-feature",
		Project:    "testproject",
		State:      workstream.StatePending,
		Objective:  longObjective,
		LastUpdate: time.Now(),
	})

	h := NewHandlers(st)
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "testproject",
			},
		},
	}
	result, err := h.HandleList(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleList() error = %v", err)
	}

	content := result.Content[0].(mcp.TextContent).Text

	// Objective should be truncated to ~100 chars + ellipsis
	if strings.Contains(content, longObjective) {
		t.Error("HandleList() should truncate long objectives, but returned full objective")
	}
	// Should contain truncated version with ellipsis
	if !strings.Contains(content, "...") {
		t.Error("HandleList() should add ellipsis to truncated objectives")
	}
	// Should be reasonably short (under 150 chars for objective field)
	if len(content) > 500 {
		// Content includes JSON structure, so total can be larger
		// but we're checking the objective isn't the full 1250 chars
		if strings.Contains(content, "objective. This is a long objective. This is a long objective. This is a long objective. This is a long objective. This is a long objective.") {
			t.Error("HandleList() objective is too long, should be truncated")
		}
	}
}

func TestHandleBugUpdate(t *testing.T) {
	st := setupTestStore(t)
	h := NewHandlers(st)

	// Add a bug first
	st.AddBug("testproject", "Feature One", "Test bug", "agent-1")
	bugs, _ := st.ListBugs(store.BugFilter{})
	bugID := bugs[0].ID

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"id":     float64(bugID),
				"status": "done",
			},
		},
	}
	result, err := h.HandleBugUpdate(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleBugUpdate() error = %v", err)
	}
	if result.IsError {
		t.Errorf("HandleBugUpdate() returned error: %v", result.Content)
	}

	// Verify bug was updated
	bugs, _ = st.ListBugs(store.BugFilter{})
	if bugs[0].Status != "done" {
		t.Errorf("Bug status = %q, want 'done'", bugs[0].Status)
	}
}

func TestHandleListPagination(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	// Create 5 workstreams
	for i := 0; i < 5; i++ {
		st.Create(&workstream.Workstream{
			Name:       "ws-" + string(rune('a'+i)),
			Project:    "proj",
			State:      workstream.StatePending,
			LastUpdate: time.Now().Add(time.Duration(i) * time.Hour),
		})
	}

	h := NewHandlers(st)

	// Get first page with limit 2
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project": "proj",
				"limit":   float64(2),
			},
		},
	}
	result, err := h.HandleList(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleList() error = %v", err)
	}

	content := result.Content[0].(mcp.TextContent).Text

	// Should have "total": 5
	if !strings.Contains(content, `"total": 5`) {
		t.Errorf("Should have total=5, got: %s", content)
	}

	// Should have next_cursor since there are more
	if !strings.Contains(content, `"next_cursor"`) {
		t.Errorf("Should have next_cursor, got: %s", content)
	}
}

func TestHandleListNameContains(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	st.Create(&workstream.Workstream{Name: "auth-login", Project: "proj", State: workstream.StatePending, LastUpdate: time.Now()})
	st.Create(&workstream.Workstream{Name: "auth-logout", Project: "proj", State: workstream.StatePending, LastUpdate: time.Now()})
	st.Create(&workstream.Workstream{Name: "api-users", Project: "proj", State: workstream.StatePending, LastUpdate: time.Now()})

	h := NewHandlers(st)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"project":       "proj",
				"name_contains": "auth",
			},
		},
	}
	result, err := h.HandleList(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleList() error = %v", err)
	}

	content := result.Content[0].(mcp.TextContent).Text

	// Should have 2 matches (auth-login and auth-logout)
	if !strings.Contains(content, `"total": 2`) {
		t.Errorf("Should have total=2 for auth search, got: %s", content)
	}
	if !strings.Contains(content, "auth-login") || !strings.Contains(content, "auth-logout") {
		t.Errorf("Should contain auth-login and auth-logout, got: %s", content)
	}
	if strings.Contains(content, "api-users") {
		t.Errorf("Should NOT contain api-users, got: %s", content)
	}
}
