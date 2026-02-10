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

// PlanItem represents a single item in the workstream plan
type PlanItem struct {
	Text     string
	Complete bool
}

// LogEntry represents a timestamped log entry
type LogEntry struct {
	Timestamp time.Time
	Content   string
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
}
