package workstream

import (
	"fmt"
	"strings"
)

const TimeFormat = "2006-01-02 15:04"

// Render converts a Workstream struct to markdown format for display
func Render(ws *Workstream) string {
	var b strings.Builder

	// Header
	b.WriteString("# Workstream: ")
	b.WriteString(ws.Name)
	b.WriteString("\n\n")

	// Status section
	b.WriteString("## Status\n")
	b.WriteString("State: ")
	b.WriteString(string(ws.State))
	b.WriteString("\n")
	b.WriteString("Last: ")
	b.WriteString(ws.LastUpdate.Format(TimeFormat))
	b.WriteString("\n")
	if ws.Owner != "" {
		b.WriteString("Owner: ")
		b.WriteString(ws.Owner)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Dependencies (only if there are any)
	if len(ws.BlockedBy) > 0 || len(ws.Blocks) > 0 {
		b.WriteString("## Dependencies\n")
		if len(ws.BlockedBy) > 0 {
			b.WriteString("Blocked by:\n")
			for _, dep := range ws.BlockedBy {
				b.WriteString(fmt.Sprintf("- %s/%s\n", dep.BlockerProject, dep.BlockerName))
			}
		}
		if len(ws.Blocks) > 0 {
			b.WriteString("Blocks:\n")
			for _, dep := range ws.Blocks {
				b.WriteString(fmt.Sprintf("- %s/%s\n", dep.BlockedProject, dep.BlockedName))
			}
		}
		b.WriteString("\n")
	}

	// Objective
	b.WriteString("## Objective\n")
	b.WriteString(ws.Objective)
	b.WriteString("\n\n")

	// Key Context
	b.WriteString("## Key Context\n")
	if ws.KeyContext != "" {
		b.WriteString(ws.KeyContext)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Plan
	b.WriteString("## Plan\n")
	for i, item := range ws.Plan {
		marker := "[ ]"
		switch item.Status {
		case TaskInProgress:
			marker = "[>]"
		case TaskDone:
			marker = "[x]"
		case TaskSkipped:
			marker = "[-]"
		default:
			// TaskPending or empty -> [ ]
			if item.Complete {
				marker = "[x]"
			}
		}
		b.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, marker, item.Text))
	}
	b.WriteString("\n")

	// Decisions
	b.WriteString("## Decisions\n")
	if ws.Decisions != "" {
		b.WriteString(ws.Decisions)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Log
	b.WriteString("## Log\n")
	for _, entry := range ws.Log {
		b.WriteString("### ")
		b.WriteString(entry.Timestamp.Format(TimeFormat))
		b.WriteString("\n")
		b.WriteString(entry.Content)
		b.WriteString("\n\n")
	}

	return b.String()
}

// Serialize is an alias for Render (for backward compatibility)
func Serialize(ws *Workstream) string {
	return Render(ws)
}
