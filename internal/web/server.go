package web

import (
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/faraz/streamctl/internal/store"
	"github.com/faraz/streamctl/pkg/workstream"
)

//go:embed templates/*.html
var templateFS embed.FS

var funcMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
}

var templates = template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"))

// Server serves the web UI for workstreams.
type Server struct {
	store   *store.Store
	project string
	mux     *http.ServeMux
}

// NewServer creates a new web server for the given project.
func NewServer(st *store.Store, project string) *Server {
	s := &Server{
		store:   st,
		project: project,
		mux:     http.NewServeMux(),
	}
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/workstream/", s.handleWorkstream)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	workstreams, err := s.store.List(store.Filter{Project: s.project})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get recent activity
	activity, err := s.store.RecentActivity(s.project, 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Compute insights
	var blocked, needsHelp, inProgress []workstream.Workstream
	for _, ws := range workstreams {
		if ws.State == workstream.StateBlocked || len(ws.BlockedBy) > 0 {
			blocked = append(blocked, ws)
		}
		if ws.NeedsHelp {
			needsHelp = append(needsHelp, ws)
		}
		if ws.State == workstream.StateInProgress {
			inProgress = append(inProgress, ws)
		}
	}

	data := struct {
		Project     string
		Workstreams []workstream.Workstream
		Activity    []workstream.ActivityEntry
		Blocked     []workstream.Workstream
		NeedsHelp   []workstream.Workstream
		InProgress  []workstream.Workstream
	}{
		Project:     s.project,
		Workstreams: workstreams,
		Activity:    activity,
		Blocked:     blocked,
		NeedsHelp:   needsHelp,
		InProgress:  inProgress,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleWorkstream(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/workstream/")
	if name == "" {
		http.NotFound(w, r)
		return
	}

	ws, err := s.store.Get(s.project, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Get all workstreams for sidebar
	allWorkstreams, err := s.store.List(store.Filter{Project: s.project})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Project        string
		Workstream     *workstream.Workstream
		AllWorkstreams []workstream.Workstream
	}{
		Project:        s.project,
		Workstream:     ws,
		AllWorkstreams: allWorkstreams,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "workstream.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
