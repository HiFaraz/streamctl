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
	Complete bool // Deprecated: use Status instead
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

	// Content sections
	Objective  string
	KeyContext string
	Plan       []PlanItem
	Decisions  string
	Log        []LogEntry

	// Dependencies
	BlockedBy []Dependency // Workstreams that block this one
	Blocks    []Dependency // Workstreams this one blocks
}
