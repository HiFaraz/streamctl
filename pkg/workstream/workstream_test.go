package workstream

import "testing"

func TestTaskStatusConstants(t *testing.T) {
	// Verify TaskStatus type and constants exist with expected values
	tests := []struct {
		status TaskStatus
		want   string
	}{
		{TaskPending, "pending"},
		{TaskInProgress, "in_progress"},
		{TaskDone, "done"},
		{TaskSkipped, "skipped"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("TaskStatus = %q, want %q", tt.status, tt.want)
		}
	}
}

func TestPlanItemHasStatus(t *testing.T) {
	item := PlanItem{
		Text:   "Test task",
		Status: TaskDone,
	}
	if item.Status != TaskDone {
		t.Errorf("PlanItem.Status = %q, want %q", item.Status, TaskDone)
	}
}

func TestWorkstreamHasDependencies(t *testing.T) {
	ws := Workstream{
		Name:    "test",
		Project: "proj",
		BlockedBy: []Dependency{
			{BlockerProject: "proj", BlockerName: "auth"},
		},
		Blocks: []Dependency{
			{BlockedProject: "proj", BlockedName: "api"},
		},
	}
	if len(ws.BlockedBy) != 1 {
		t.Errorf("BlockedBy len = %d, want 1", len(ws.BlockedBy))
	}
	if len(ws.Blocks) != 1 {
		t.Errorf("Blocks len = %d, want 1", len(ws.Blocks))
	}
}
