package workstream

import (
	"strings"
	"testing"
	"time"
)

func TestRenderBasic(t *testing.T) {
	ws := &Workstream{
		Name:       "Test Workstream",
		State:      StateInProgress,
		LastUpdate: time.Date(2026, 2, 10, 14, 30, 0, 0, time.UTC),
		Objective:  "Build something cool.",
		Plan: []PlanItem{
			{Text: "Step one", Complete: false},
			{Text: "Step two", Complete: true},
		},
	}

	output := Render(ws)

	// Verify key parts are present
	if !strings.Contains(output, "# Workstream: Test Workstream") {
		t.Errorf("Missing workstream name header")
	}
	if !strings.Contains(output, "State: in_progress") {
		t.Errorf("Missing state")
	}
	if !strings.Contains(output, "Last: 2026-02-10 14:30") {
		t.Errorf("Missing last update")
	}
	if !strings.Contains(output, "Build something cool.") {
		t.Errorf("Missing objective")
	}
	if !strings.Contains(output, "1. [ ] Step one") {
		t.Errorf("Missing incomplete plan item")
	}
	if !strings.Contains(output, "2. [x] Step two") {
		t.Errorf("Missing complete plan item")
	}
}

func TestRenderWithOwner(t *testing.T) {
	ws := &Workstream{
		Name:       "Owned Workstream",
		State:      StateInProgress,
		LastUpdate: time.Date(2026, 2, 10, 14, 30, 0, 0, time.UTC),
		Owner:      "agent-123",
		Objective:  "Test owner field.",
	}

	output := Render(ws)

	if !strings.Contains(output, "Owner: agent-123") {
		t.Errorf("Missing owner field")
	}
}

func TestRenderWithLog(t *testing.T) {
	ws := &Workstream{
		Name:       "Logged Workstream",
		State:      StatePending,
		LastUpdate: time.Date(2026, 2, 10, 14, 30, 0, 0, time.UTC),
		Objective:  "Test log entries.",
		Log: []LogEntry{
			{
				Timestamp: time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC),
				Content:   "First log entry.",
			},
			{
				Timestamp: time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC),
				Content:   "Second log entry.",
			},
		},
	}

	output := Render(ws)

	if !strings.Contains(output, "### 2026-02-10 10:00") {
		t.Errorf("Missing first log timestamp")
	}
	if !strings.Contains(output, "First log entry.") {
		t.Errorf("Missing first log content")
	}
	if !strings.Contains(output, "### 2026-02-10 12:00") {
		t.Errorf("Missing second log timestamp")
	}
}

func TestRenderAllStates(t *testing.T) {
	states := []State{StatePending, StateInProgress, StateBlocked, StateDone}

	for _, state := range states {
		ws := &Workstream{
			Name:       "State Test",
			State:      state,
			LastUpdate: time.Now(),
			Objective:  "Test state rendering.",
		}

		output := Render(ws)

		if !strings.Contains(output, "State: "+string(state)) {
			t.Errorf("State %q not rendered correctly", state)
		}
	}
}
