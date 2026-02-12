package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/faraz/streamctl/internal/store"
	"github.com/faraz/streamctl/pkg/workstream"
)

// View represents the current view state
type View int

const (
	ProjectListView View = iota
	WorkstreamListView
	DetailView
)

// KeyMap defines key bindings
type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Quit    key.Binding
	Refresh key.Binding
}

var keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}

// Model is the main TUI model
type Model struct {
	store           *store.Store
	view            View
	projects        []string
	workstreams     []workstream.Workstream
	selected        *workstream.Workstream
	cursor          int
	currentProject  string
	width           int
	height          int
	err             error
}

// New creates a new TUI model
func New(st *store.Store) Model {
	return Model{
		store: st,
		view:  ProjectListView,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.loadProjects
}

func (m Model) loadProjects() tea.Msg {
	projects, err := m.store.ListProjects()
	if err != nil {
		return errMsg{err}
	}
	return projectsMsg(projects)
}

func (m Model) loadWorkstreams() tea.Msg {
	workstreams, err := m.store.List(store.Filter{Project: m.currentProject})
	if err != nil {
		return errMsg{err}
	}
	return workstreamsMsg(workstreams)
}

func (m Model) loadDetail() tea.Msg {
	if m.cursor >= len(m.workstreams) {
		return nil
	}
	ws := m.workstreams[m.cursor]
	detail, err := m.store.Get(ws.Project, ws.Name)
	if err != nil {
		return errMsg{err}
	}
	return detailMsg{detail}
}

type errMsg struct{ err error }
type projectsMsg []string
type workstreamsMsg []workstream.Workstream
type detailMsg struct{ ws *workstream.Workstream }

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case key.Matches(msg, keys.Down):
			maxItems := m.maxItems()
			if m.cursor < maxItems-1 {
				m.cursor++
			}
			return m, nil

		case key.Matches(msg, keys.Enter):
			return m.handleSelect()

		case key.Matches(msg, keys.Back):
			return m.handleBack()

		case key.Matches(msg, keys.Refresh):
			return m.refresh()
		}

	case errMsg:
		m.err = msg.err
		return m, nil

	case projectsMsg:
		m.projects = msg
		m.cursor = 0
		return m, nil

	case workstreamsMsg:
		m.workstreams = msg
		m.cursor = 0
		return m, nil

	case detailMsg:
		m.selected = msg.ws
		m.view = DetailView
		return m, nil
	}

	return m, nil
}

func (m Model) maxItems() int {
	switch m.view {
	case ProjectListView:
		return len(m.projects)
	case WorkstreamListView:
		return len(m.workstreams)
	default:
		return 0
	}
}

func (m Model) handleSelect() (tea.Model, tea.Cmd) {
	switch m.view {
	case ProjectListView:
		if m.cursor < len(m.projects) {
			m.currentProject = m.projects[m.cursor]
			m.view = WorkstreamListView
			m.cursor = 0
			return m, m.loadWorkstreams
		}
	case WorkstreamListView:
		if m.cursor < len(m.workstreams) {
			return m, m.loadDetail
		}
	}
	return m, nil
}

func (m Model) handleBack() (tea.Model, tea.Cmd) {
	switch m.view {
	case WorkstreamListView:
		m.view = ProjectListView
		m.cursor = 0
		m.currentProject = ""
	case DetailView:
		m.view = WorkstreamListView
		m.selected = nil
	}
	return m, nil
}

func (m Model) refresh() (tea.Model, tea.Cmd) {
	switch m.view {
	case ProjectListView:
		return m, m.loadProjects
	case WorkstreamListView:
		return m, m.loadWorkstreams
	case DetailView:
		return m, m.loadDetail
	}
	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	var s strings.Builder

	// Header
	s.WriteString(m.renderHeader())
	s.WriteString("\n\n")

	// Content
	switch m.view {
	case ProjectListView:
		s.WriteString(m.renderProjectList())
	case WorkstreamListView:
		s.WriteString(m.renderWorkstreamList())
	case DetailView:
		s.WriteString(m.renderDetail())
	}

	// Footer
	s.WriteString("\n\n")
	s.WriteString(m.renderHelp())

	return s.String()
}

func (m Model) renderHeader() string {
	var title string
	switch m.view {
	case ProjectListView:
		title = "Workstreams"
	case WorkstreamListView:
		title = fmt.Sprintf("Workstreams > %s", m.currentProject)
	case DetailView:
		if m.selected != nil {
			title = fmt.Sprintf("Workstreams > %s > %s", m.currentProject, m.selected.Name)
		}
	}
	return headerStyle.Render(title)
}

func (m Model) renderProjectList() string {
	if len(m.projects) == 0 {
		return subtitleStyle.Render("No projects found")
	}

	var s strings.Builder
	for i, p := range m.projects {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "> "
			style = selectedStyle
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(p)))
	}
	return s.String()
}

func (m Model) renderWorkstreamList() string {
	if len(m.workstreams) == 0 {
		return subtitleStyle.Render("No workstreams found")
	}

	var s strings.Builder
	for i, ws := range m.workstreams {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "> "
			style = selectedStyle
		}

		stateStr := StateStyle(string(ws.State)).Render(fmt.Sprintf("[%s]", ws.State))
		ownerStr := ""
		if ws.Owner != "" {
			ownerStr = subtitleStyle.Render(fmt.Sprintf(" (%s)", ws.Owner))
		}

		line := fmt.Sprintf("%s %s%s", style.Render(ws.Name), stateStr, ownerStr)
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, line))
	}
	return s.String()
}

func (m Model) renderDetail() string {
	if m.selected == nil {
		return "No workstream selected"
	}

	ws := m.selected
	var s strings.Builder

	// State badge
	stateStyle := StateStyle(string(ws.State))
	s.WriteString(fmt.Sprintf("State: %s\n", stateStyle.Render(string(ws.State))))
	s.WriteString(fmt.Sprintf("Last:  %s\n", ws.LastUpdate.Format("2006-01-02 15:04")))
	if ws.Owner != "" {
		s.WriteString(fmt.Sprintf("Owner: %s\n", ws.Owner))
	}
	s.WriteString("\n")

	// Objective
	s.WriteString(titleStyle.Render("Objective"))
	s.WriteString("\n")
	s.WriteString(ws.Objective)
	s.WriteString("\n\n")

	// Plan
	if len(ws.Plan) > 0 {
		s.WriteString(titleStyle.Render("Plan"))
		s.WriteString("\n")
		for i, item := range ws.Plan {
			check := "[ ]"
			if item.Complete {
				check = "[x]"
			}
			s.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, check, item.Text))
		}
		s.WriteString("\n")
	}

	// Recent log entries (last 3)
	if len(ws.Log) > 0 {
		s.WriteString(titleStyle.Render("Recent Log"))
		s.WriteString("\n")
		start := 0
		if len(ws.Log) > 3 {
			start = len(ws.Log) - 3
		}
		for _, entry := range ws.Log[start:] {
			s.WriteString(subtitleStyle.Render(entry.Timestamp.Format("2006-01-02 15:04")))
			s.WriteString("\n")
			// Truncate long entries
			content := entry.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			s.WriteString(content)
			s.WriteString("\n\n")
		}
	}

	return s.String()
}

func (m Model) renderHelp() string {
	var help string
	switch m.view {
	case ProjectListView:
		help = "[↑/k] Up  [↓/j] Down  [Enter] Select  [r] Refresh  [q] Quit"
	case WorkstreamListView:
		help = "[↑/k] Up  [↓/j] Down  [Enter] Select  [Esc] Back  [r] Refresh  [q] Quit"
	case DetailView:
		help = "[Esc] Back  [r] Refresh  [q] Quit"
	}
	return helpStyle.Render(help)
}

// Run starts the TUI
func Run(st *store.Store) error {
	p := tea.NewProgram(New(st), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
