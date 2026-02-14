package web

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/faraz/streamctl/internal/store"
	"github.com/faraz/streamctl/pkg/workstream"
)

//go:embed templates/*.html
var templateFS embed.FS

var funcMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"json": func(v any) template.JS {
		b, _ := json.Marshal(v)
		return template.JS(b)
	},
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
	s.mux.HandleFunc("/search", s.handleSearch)
	s.mux.HandleFunc("/api/activity", s.handleActivityAPI)
	s.mux.HandleFunc("/api/search", s.handleSearchAPI)
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

	// Get recent activity (fetch one extra to check if there's more)
	const pageSize = 20
	activity, err := s.store.RecentActivity(s.project, pageSize+1, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasMore := len(activity) > pageSize
	if hasMore {
		activity = activity[:pageSize]
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
		HasMore     bool
	}{
		Project:     s.project,
		Workstreams: workstreams,
		Activity:    activity,
		Blocked:     blocked,
		NeedsHelp:   needsHelp,
		InProgress:  inProgress,
		HasMore:     hasMore,
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

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	workstreams, err := s.store.List(store.Filter{Project: s.project})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Project     string
		Workstreams []workstream.Workstream
	}{
		Project:     s.project,
		Workstreams: workstreams,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "search.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleActivityAPI(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	activity, err := s.store.RecentActivity(s.project, limit+1, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasMore := len(activity) > limit
	if hasMore {
		activity = activity[:limit]
	}

	// Convert to JSON-friendly format
	type jsonEntry struct {
		WorkstreamName string `json:"workstreamName"`
		Timestamp      int64  `json:"timestamp"`
		Content        string `json:"content"`
		NeedsHelp      bool   `json:"needsHelp"`
		BlockedBy      string `json:"blockedBy,omitempty"`
		RelativeTime   string `json:"relativeTime"`
	}

	entries := make([]jsonEntry, len(activity))
	for i, e := range activity {
		entries[i] = jsonEntry{
			WorkstreamName: e.WorkstreamName,
			Timestamp:      e.Timestamp.Unix(),
			Content:        e.Content,
			NeedsHelp:      e.NeedsHelp,
			BlockedBy:      e.BlockedBy,
			RelativeTime:   e.RelativeTime,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"entries": entries,
		"hasMore": hasMore,
	})
}

func (s *Server) handleSearchAPI(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	wsFilter := r.URL.Query().Get("ws")

	results, err := s.store.Search(s.project, query, wsFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
