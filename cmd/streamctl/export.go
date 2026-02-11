package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/faraz/streamctl/internal/store"
	"github.com/faraz/streamctl/pkg/workstream"
)

const generatedHeader = `<!-- GENERATED FILE - DO NOT EDIT -->
<!-- Source: streamctl database -->
<!-- Regenerate with: streamctl export %s/%s -->

`

// exportWorkstream exports a single workstream to the given writer as markdown.
func exportWorkstream(s *store.Store, project, name string, w io.Writer) error {
	ws, err := s.Get(project, name)
	if err != nil {
		return fmt.Errorf("workstream not found: %s/%s", project, name)
	}

	header := fmt.Sprintf(generatedHeader, project, name)
	if _, err := w.Write([]byte(header)); err != nil {
		return err
	}
	_, err = w.Write([]byte(workstream.Render(ws)))
	return err
}

// exportAllWorkstreams exports all workstreams for a project to the given directory.
func exportAllWorkstreams(s *store.Store, project, dir string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// List all workstreams for this project
	workstreams, err := s.List(store.Filter{Project: project})
	if err != nil {
		return fmt.Errorf("failed to list workstreams: %w", err)
	}

	// Export each workstream
	for _, ws := range workstreams {
		path := filepath.Join(dir, ws.Name+".md")
		header := fmt.Sprintf(generatedHeader, ws.Project, ws.Name)
		content := header + workstream.Render(&ws)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	return nil
}
