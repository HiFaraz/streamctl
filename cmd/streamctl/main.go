package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/faraz/streamctl/internal/mcp"
	"github.com/faraz/streamctl/internal/store"
	"github.com/faraz/streamctl/internal/tui"
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
	case "tui":
		st := mustOpenStore(dbPath)
		defer st.Close()
		runTUI(st)
	case "list":
		st := mustOpenStore(dbPath)
		defer st.Close()
		runList(st)
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
  streamctl init               Initialize the database
  streamctl serve              Start MCP server (stdio)
  streamctl tui                Launch TUI dashboard
  streamctl list [--project X] List workstreams (JSON)
  streamctl version            Show version
  streamctl help               Show this help

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

func runTUI(st *store.Store) {
	if err := tui.Run(st); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
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
