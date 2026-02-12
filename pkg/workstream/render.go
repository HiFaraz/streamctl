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
		if item.Notes != "" {
			// Indent notes under the task
			lines := strings.Split(item.Notes, "\n")
			for _, line := range lines {
				b.WriteString("   ")
				b.WriteString(line)
				b.WriteString("\n")
			}
		}
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

// RenderMilestone converts a Milestone to markdown format
func RenderMilestone(m *Milestone) string {
	var b strings.Builder

	b.WriteString("# Milestone: ")
	b.WriteString(m.Name)
	b.WriteString("\n\n")

	b.WriteString("## Status\n")
	b.WriteString("State: ")
	b.WriteString(string(m.Status))
	b.WriteString("\n")
	if !m.CreatedAt.IsZero() {
		b.WriteString("Created: ")
		b.WriteString(m.CreatedAt.Format(TimeFormat))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if m.Description != "" {
		b.WriteString("## Description\n")
		b.WriteString(m.Description)
		b.WriteString("\n\n")
	}

	b.WriteString("## Requirements\n")
	if len(m.Requirements) == 0 {
		b.WriteString("_No requirements defined_\n")
	} else {
		doneCount := 0
		for _, req := range m.Requirements {
			marker := "[ ]"
			if req.WorkstreamState == StateDone {
				marker = "[x]"
				doneCount++
			}
			b.WriteString(fmt.Sprintf("- %s %s/%s (%s)\n", marker, req.WorkstreamProject, req.WorkstreamName, req.WorkstreamState))
		}
		b.WriteString(fmt.Sprintf("\nProgress: %d/%d complete\n", doneCount, len(m.Requirements)))
	}

	return b.String()
}
