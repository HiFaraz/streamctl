package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/faraz/streamctl/internal/mcp"
	"github.com/faraz/streamctl/internal/store"
	"github.com/faraz/streamctl/internal/web"
	"github.com/mark3labs/mcp-go/server"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Get database path
	// Priority: STREAMCTL_DB env > .streamctl/workstreams.db in cwd > ~/.streamctl/workstreams.db
	dbPath := os.Getenv("STREAMCTL_DB")
	if dbPath == "" {
		// Check for project-local database
		if _, err := os.Stat(".streamctl/workstreams.db"); err == nil {
			dbPath = ".streamctl/workstreams.db"
		} else {
			home, _ := os.UserHomeDir()
			dbPath = filepath.Join(home, ".streamctl", "workstreams.db")
		}
	}

	switch os.Args[1] {
	case "init":
		runInit(dbPath)
	case "serve":
		st := mustOpenStore(dbPath)
		defer st.Close()
		runServer(st)
	case "list":
		st := mustOpenStore(dbPath)
		defer st.Close()
		runList(st)
	case "web":
		st := mustOpenStore(dbPath)
		defer st.Close()
		runWeb(st)
	case "export":
		st := mustOpenStore(dbPath)
		defer st.Close()
		runExport(st)
	case "version", "--version", "-v":
		fmt.Println("streamctl", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`streamctl - Manage workstreams across projects

Usage:
  streamctl init                        Initialize the database
  streamctl serve                       Start MCP server (stdio)
  streamctl web [--port PORT]           Start web UI (default: 8080)
  streamctl list [--project X]          List workstreams (JSON)
  streamctl export PROJECT/NAME         Export single workstream to stdout
  streamctl export PROJECT [--dir DIR]  Export all workstreams to directory
  streamctl version                     Show version
  streamctl help                        Show this help

Database location (in priority order):
  1. STREAMCTL_DB env variable
  2. .streamctl/workstreams.db (project-local)
  3. ~/.streamctl/workstreams.db (user global)`)
}

func mustOpenStore(dbPath string) *store.Store {
	st, err := store.New(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'streamctl init' to create the database.\n")
		os.Exit(1)
	}
	return st
}

func runInit(dbPath string) {
	// Create directory if needed
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Create/open database (migrations run automatically)
	st, err := store.New(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}
	st.Close()

	fmt.Printf("Initialized database at %s\n", dbPath)
}

func runServer(st *store.Store) {
	s := mcp.NewServer(st)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func runList(st *store.Store) {
	filter := store.Filter{}

	// Parse --project flag
	for i, arg := range os.Args[2:] {
		if arg == "--project" && i+1 < len(os.Args[2:]) {
			filter.Project = os.Args[i+3]
		}
	}

	workstreams, err := st.List(filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Output as JSON
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(workstreams)
}

func runExport(st *store.Store) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: streamctl export PROJECT/NAME or streamctl export PROJECT [--dir DIR]")
		os.Exit(1)
	}

	arg := os.Args[2]

	// Check if it's PROJECT/NAME format
	if idx := indexOf(arg, '/'); idx != -1 {
		project := arg[:idx]
		name := arg[idx+1:]
		if err := exportWorkstream(st, project, name, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Otherwise it's PROJECT [--dir DIR]
	project := arg
	dir := "./workstreams"

	for i, a := range os.Args[3:] {
		if a == "--dir" && i+1 < len(os.Args[3:]) {
			dir = os.Args[i+4]
		}
	}

	if err := exportAllWorkstreams(st, project, dir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exported workstreams to %s/\n", dir)
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func runWeb(st *store.Store) {
	port := "8080"

	// Parse --port flag
	for i, arg := range os.Args[2:] {
		if arg == "--port" && i+1 < len(os.Args[2:]) {
			port = os.Args[i+3]
		}
	}

	// Detect project from current directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	project, err := web.DetectProject(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting project: %v\n", err)
		os.Exit(1)
	}

	srv := web.NewServer(st, project)

	fmt.Printf("Serving %s workstreams at http://localhost:%s\n", project, port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

