package mcp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/faraz/streamctl/internal/store"
	"github.com/faraz/streamctl/pkg/workstream"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Handlers provides MCP tool handlers
type Handlers struct {
	store *store.Store
}

// NewHandlers creates a new Handlers instance
func NewHandlers(st *store.Store) *Handlers {
	return &Handlers{store: st}
}

// RegisterTools registers all workstream tools with the MCP server
func (h *Handlers) RegisterTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("workstream_list",
			mcp.WithDescription("List workstreams with optional filters"),
			mcp.WithString("project", mcp.Description("Filter by project name")),
			mcp.WithString("state", mcp.Description("Filter by state: pending, in_progress, blocked, done")),
			mcp.WithString("owner", mcp.Description("Filter by owner")),
		),
		h.HandleList,
	)

	s.AddTool(
		mcp.NewTool("workstream_get",
			mcp.WithDescription("Get full workstream content"),
			mcp.WithString("project", mcp.Description("Project name"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Workstream name (without .md)"), mcp.Required()),
		),
		h.HandleGet,
	)

	s.AddTool(
		mcp.NewTool("workstream_create",
			mcp.WithDescription("Create a new workstream from template"),
			mcp.WithString("project", mcp.Description("Project name"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Workstream name (without .md)"), mcp.Required()),
			mcp.WithString("objective", mcp.Description("One-sentence objective"), mcp.Required()),
		),
		h.HandleCreate,
	)

	s.AddTool(
		mcp.NewTool("workstream_update",
			mcp.WithDescription("Update workstream fields"),
			mcp.WithString("project", mcp.Description("Project name"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Workstream name"), mcp.Required()),
			mcp.WithString("state", mcp.Description("New state: pending, in_progress, blocked, done")),
			mcp.WithString("log_entry", mcp.Description("New log entry to append")),
			mcp.WithNumber("plan_index", mcp.Description("Toggle completion of plan item at this index")),
			mcp.WithString("task_add", mcp.Description("Add a new task with this text")),
			mcp.WithNumber("task_remove", mcp.Description("Remove task at this position (0-indexed)")),
			mcp.WithObject("task_status", mcp.Description("Set task status: {\"position\": 0, \"status\": \"done\"}")),
			mcp.WithString("add_blocker", mcp.Description("Add dependency: 'project/workstream' blocks this one")),
			mcp.WithString("remove_blocker", mcp.Description("Remove dependency from this workstream")),
		),
		h.HandleUpdate,
	)

	s.AddTool(
		mcp.NewTool("workstream_claim",
			mcp.WithDescription("Set ownership of a workstream"),
			mcp.WithString("project", mcp.Description("Project name"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Workstream name"), mcp.Required()),
			mcp.WithString("owner", mcp.Description("Owner identifier"), mcp.Required()),
		),
		h.HandleClaim,
	)

	s.AddTool(
		mcp.NewTool("workstream_release",
			mcp.WithDescription("Clear ownership of a workstream"),
			mcp.WithString("project", mcp.Description("Project name"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Workstream name"), mcp.Required()),
		),
		h.HandleRelease,
	)
}

// HandleList lists workstreams with optional filters
func (h *Handlers) HandleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project := mcp.ParseString(req, "project", "")
	state := mcp.ParseString(req, "state", "")
	owner := mcp.ParseString(req, "owner", "")

	filter := store.Filter{
		Project: project,
		State:   workstream.State(state),
		Owner:   owner,
	}

	workstreams, err := h.store.List(filter)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Convert to summary format
	type wsSummary struct {
		Project    string `json:"project"`
		Name       string `json:"name"`
		State      string `json:"state"`
		LastUpdate string `json:"last_update"`
		Owner      string `json:"owner,omitempty"`
		Objective  string `json:"objective"`
	}

	summaries := make([]wsSummary, len(workstreams))
	for i, ws := range workstreams {
		summaries[i] = wsSummary{
			Project:    ws.Project,
			Name:       ws.Name,
			State:      string(ws.State),
			LastUpdate: ws.LastUpdate.Format("2006-01-02 15:04"),
			Owner:      ws.Owner,
			Objective:  ws.Objective,
		}
	}

	data, _ := json.MarshalIndent(summaries, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// HandleGet returns a single workstream
func (h *Handlers) HandleGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project := mcp.ParseString(req, "project", "")
	name := mcp.ParseString(req, "name", "")

	if project == "" || name == "" {
		return mcp.NewToolResultError("project and name are required"), nil
	}

	ws, err := h.store.Get(project, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Return as markdown
	return mcp.NewToolResultText(workstream.Serialize(ws)), nil
}

// HandleCreate creates a new workstream
func (h *Handlers) HandleCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project := mcp.ParseString(req, "project", "")
	name := mcp.ParseString(req, "name", "")
	objective := mcp.ParseString(req, "objective", "")

	if project == "" || name == "" || objective == "" {
		return mcp.NewToolResultError("project, name, and objective are required"), nil
	}

	ws := &workstream.Workstream{
		Name:       name,
		Project:    project,
		State:      workstream.StatePending,
		LastUpdate: time.Now().UTC().Truncate(time.Minute),
		Objective:  objective,
	}

	if err := h.store.Create(ws); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("Created workstream: " + project + "/" + name), nil
}

// HandleUpdate updates a workstream
func (h *Handlers) HandleUpdate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project := mcp.ParseString(req, "project", "")
	name := mcp.ParseString(req, "name", "")

	if project == "" || name == "" {
		return mcp.NewToolResultError("project and name are required"), nil
	}

	var updates store.WorkstreamUpdate

	if state := mcp.ParseString(req, "state", ""); state != "" {
		s := workstream.State(state)
		updates.State = &s
	}

	if logEntry := mcp.ParseString(req, "log_entry", ""); logEntry != "" {
		updates.LogEntry = &logEntry
	}

	args, _ := req.Params.Arguments.(map[string]any)

	// Check if plan_index was provided
	if args != nil {
		if _, ok := args["plan_index"]; ok {
			idx := mcp.ParseInt(req, "plan_index", -1)
			if idx >= 0 {
				updates.PlanIndex = &idx
			}
		}
	}

	// Handle task_add
	if taskAdd := mcp.ParseString(req, "task_add", ""); taskAdd != "" {
		if err := h.store.AddTask(project, name, taskAdd); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	// Handle task_remove
	if args != nil {
		if _, ok := args["task_remove"]; ok {
			idx := mcp.ParseInt(req, "task_remove", -1)
			if idx >= 0 {
				if err := h.store.RemoveTask(project, name, idx); err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
			}
		}
	}

	// Handle task_status
	if args != nil {
		if statusObj, ok := args["task_status"].(map[string]any); ok {
			position := int(statusObj["position"].(float64))
			status := workstream.TaskStatus(statusObj["status"].(string))
			if err := h.store.SetTaskStatus(project, name, position, status); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
		}
	}

	// Handle add_blocker (format: "project/name")
	if addBlocker := mcp.ParseString(req, "add_blocker", ""); addBlocker != "" {
		parts := splitProjectName(addBlocker)
		if len(parts) == 2 {
			if err := h.store.AddDependency(parts[0], parts[1], project, name); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
		} else {
			return mcp.NewToolResultError("add_blocker must be in format 'project/name'"), nil
		}
	}

	// Handle remove_blocker (format: "project/name")
	if removeBlocker := mcp.ParseString(req, "remove_blocker", ""); removeBlocker != "" {
		parts := splitProjectName(removeBlocker)
		if len(parts) == 2 {
			if err := h.store.RemoveDependency(parts[0], parts[1], project, name); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
		} else {
			return mcp.NewToolResultError("remove_blocker must be in format 'project/name'"), nil
		}
	}

	if err := h.store.Update(project, name, updates); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("Updated workstream: " + project + "/" + name), nil
}

// splitProjectName splits "project/name" into parts
func splitProjectName(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// HandleClaim sets ownership of a workstream
func (h *Handlers) HandleClaim(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project := mcp.ParseString(req, "project", "")
	name := mcp.ParseString(req, "name", "")
	owner := mcp.ParseString(req, "owner", "")

	if project == "" || name == "" || owner == "" {
		return mcp.NewToolResultError("project, name, and owner are required"), nil
	}

	updates := store.WorkstreamUpdate{
		Owner: &owner,
	}

	if err := h.store.Update(project, name, updates); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("Claimed workstream: " + project + "/" + name + " for " + owner), nil
}

// HandleRelease clears ownership of a workstream
func (h *Handlers) HandleRelease(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project := mcp.ParseString(req, "project", "")
	name := mcp.ParseString(req, "name", "")

	if project == "" || name == "" {
		return mcp.NewToolResultError("project and name are required"), nil
	}

	emptyOwner := ""
	updates := store.WorkstreamUpdate{
		Owner: &emptyOwner,
	}

	if err := h.store.Update(project, name, updates); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("Released workstream: " + project + "/" + name), nil
}

// NewServer creates a new MCP server with workstream tools
func NewServer(st *store.Store) *server.MCPServer {
	s := server.NewMCPServer("workstreams", "1.0.0")
	h := NewHandlers(st)
	h.RegisterTools(s)
	return s
}
