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

func TestRenderTaskStatus(t *testing.T) {
	ws := &Workstream{
		Name:       "Task Status Test",
		State:      StateInProgress,
		LastUpdate: time.Now(),
		Objective:  "Test task status markers.",
		Plan: []PlanItem{
			{Text: "Pending task", Status: TaskPending},
			{Text: "In progress task", Status: TaskInProgress},
			{Text: "Done task", Status: TaskDone},
			{Text: "Skipped task", Status: TaskSkipped},
		},
	}

	output := Render(ws)

	// [ ] for pending
	if !strings.Contains(output, "1. [ ] Pending task") {
		t.Errorf("Pending task not rendered with [ ] marker")
	}
	// [>] for in_progress
	if !strings.Contains(output, "2. [>] In progress task") {
		t.Errorf("In progress task not rendered with [>] marker")
	}
	// [x] for done
	if !strings.Contains(output, "3. [x] Done task") {
		t.Errorf("Done task not rendered with [x] marker")
	}
	// [-] for skipped
	if !strings.Contains(output, "4. [-] Skipped task") {
		t.Errorf("Skipped task not rendered with [-] marker")
	}
}

func TestRenderTaskNotes(t *testing.T) {
	ws := &Workstream{
		Name:       "Notes Test",
		State:      StateInProgress,
		LastUpdate: time.Now(),
		Objective:  "Test notes rendering.",
		Plan: []PlanItem{
			{Text: "Task with notes", Status: TaskInProgress, Notes: "## Details\n- Item 1\n- Item 2"},
			{Text: "Task without notes", Status: TaskPending},
		},
	}

	output := Render(ws)

	// Notes should be indented under the task
	if !strings.Contains(output, "1. [>] Task with notes\n   ## Details") {
		t.Errorf("Notes not rendered correctly under task")
	}
	if !strings.Contains(output, "   - Item 1") {
		t.Errorf("Notes lines not indented")
	}
	// Task without notes should not have extra indented lines
	if strings.Contains(output, "2. [ ] Task without notes\n   ") {
		t.Errorf("Empty notes should not add indented lines")
	}
}

func TestRenderDependencies(t *testing.T) {
	ws := &Workstream{
		Name:       "Dependency Test",
		State:      StateBlocked,
		LastUpdate: time.Now(),
		Objective:  "Test dependencies.",
		BlockedBy: []Dependency{
			{BlockerProject: "proj", BlockerName: "auth"},
			{BlockerProject: "proj", BlockerName: "core"},
		},
		Blocks: []Dependency{
			{BlockedProject: "proj", BlockedName: "api"},
		},
	}

	output := Render(ws)

	if !strings.Contains(output, "## Dependencies") {
		t.Errorf("Missing Dependencies section")
	}
	if !strings.Contains(output, "Blocked by:") {
		t.Errorf("Missing 'Blocked by:' label")
	}
	if !strings.Contains(output, "- proj/auth") {
		t.Errorf("Missing blocker proj/auth")
	}
	if !strings.Contains(output, "- proj/core") {
		t.Errorf("Missing blocker proj/core")
	}
	if !strings.Contains(output, "Blocks:") {
		t.Errorf("Missing 'Blocks:' label")
	}
	if !strings.Contains(output, "- proj/api") {
		t.Errorf("Missing blocked proj/api")
	}
}

func TestRenderNoDependencies(t *testing.T) {
	ws := &Workstream{
		Name:       "No Deps Test",
		State:      StatePending,
		LastUpdate: time.Now(),
		Objective:  "Test no dependencies.",
	}

	output := Render(ws)

	// Should NOT have Dependencies section when empty
	if strings.Contains(output, "## Dependencies") {
		t.Errorf("Dependencies section should be omitted when empty")
	}
}
