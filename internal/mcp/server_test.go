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
