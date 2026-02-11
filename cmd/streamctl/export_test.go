package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faraz/streamctl/internal/store"
	"github.com/faraz/streamctl/pkg/workstream"
)

func TestExportSingleWorkstream(t *testing.T) {
	// Setup: create a temp database with a workstream
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	err = s.Create(&workstream.Workstream{
		Project:   "myproject",
		Name:      "auth",
		Objective: "Implement authentication",
	})
	if err != nil {
		t.Fatalf("failed to create workstream: %v", err)
	}

	// Export to stdout (capture output)
	var buf strings.Builder
	err = exportWorkstream(s, "myproject", "auth", &buf)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "GENERATED FILE - DO NOT EDIT") {
		t.Errorf("expected generated header, got: %s", output)
	}
	if !strings.Contains(output, "# Workstream: auth") {
		t.Errorf("expected markdown header, got: %s", output)
	}
	if !strings.Contains(output, "Implement authentication") {
		t.Errorf("expected objective in output, got: %s", output)
	}
}

func TestExportAllWorkstreams(t *testing.T) {
	// Setup
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	s.Create(&workstream.Workstream{Project: "myproject", Name: "auth", Objective: "Implement authentication"})
	s.Create(&workstream.Workstream{Project: "myproject", Name: "api", Objective: "Build REST API"})
	s.Create(&workstream.Workstream{Project: "otherproject", Name: "unrelated", Objective: "Should not be exported"})

	// Export to directory
	outDir := filepath.Join(t.TempDir(), "workstreams")
	err = exportAllWorkstreams(s, "myproject", outDir)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Check files were created
	authPath := filepath.Join(outDir, "auth.md")
	if _, err := os.Stat(authPath); os.IsNotExist(err) {
		t.Errorf("expected auth.md to exist")
	}

	apiPath := filepath.Join(outDir, "api.md")
	if _, err := os.Stat(apiPath); os.IsNotExist(err) {
		t.Errorf("expected api.md to exist")
	}

	// Check unrelated project not exported
	unrelatedPath := filepath.Join(outDir, "unrelated.md")
	if _, err := os.Stat(unrelatedPath); !os.IsNotExist(err) {
		t.Errorf("unrelated.md should not exist")
	}

	// Check content
	content, _ := os.ReadFile(authPath)
	if !strings.Contains(string(content), "GENERATED FILE - DO NOT EDIT") {
		t.Errorf("expected generated header in auth.md")
	}
	if !strings.Contains(string(content), "# Workstream: auth") {
		t.Errorf("expected markdown header in auth.md")
	}
}
