package workstream

import "time"

// State represents the current state of a workstream
type State string

const (
	StatePending    State = "pending"
	StateInProgress State = "in_progress"
	StateBlocked    State = "blocked"
	StateDone       State = "done"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskDone       TaskStatus = "done"
	TaskSkipped    TaskStatus = "skipped"
)

// PlanItem represents a single item in the workstream plan
type PlanItem struct {
	Text     string
	Status   TaskStatus
	Notes    string // Markdown-formatted notes (code snippets, details, links)
	Complete bool   // Deprecated: use Status instead
}

// LogEntry represents a timestamped log entry
type LogEntry struct {
	Timestamp time.Time
	Content   string
}

// Dependency represents a blocking relationship between workstreams
type Dependency struct {
	BlockerProject string
	BlockerName    string
	BlockedProject string
	BlockedName    string
}

// Workstream represents a parsed workstream markdown file
type Workstream struct {
	Name       string // From H1: "# Workstream: NAME"
	Project    string // Directory name (e.g., "fleetadm")
	FilePath   string // Full path to .md file

	// Status section
	State      State
	LastUpdate time.Time
	Owner      string // Optional
	NeedsHelp  bool   // Flag indicating workstream is stuck/at-risk

	// Content sections
	Objective string
	Plan      []PlanItem
	Log       []LogEntry

	// Dependencies
	BlockedBy []Dependency // Workstreams that block this one
	Blocks    []Dependency // Workstreams this one blocks
}

// ActivityEntry represents a log entry with workstream context
type ActivityEntry struct {
	WorkstreamName    string
	WorkstreamProject string
	Timestamp         time.Time
	Content           string
	NeedsHelp         bool   // Workstream needs help flag
	BlockedBy         string // First blocker name if blocked
	RelativeTime      string // Human-readable relative time
}

// Milestone represents a cross-workstream gate/checkpoint
type Milestone struct {
	Name         string
	Project      string
	Description  string
	CreatedAt    time.Time
	Status       State // Computed: pending/in_progress/done
	Requirements []MilestoneRequirement
}

// MilestoneRequirement represents a workstream required for a milestone
type MilestoneRequirement struct {
	WorkstreamProject string
	WorkstreamName    string
	WorkstreamState   State // Current state of the workstream
}
