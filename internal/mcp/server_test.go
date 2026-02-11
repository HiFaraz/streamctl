package mcp

import (
	"context"
	"path/filepath"
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
