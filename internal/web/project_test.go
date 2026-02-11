package web

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDetectProject_FromGitRemote(t *testing.T) {
	// Create a temp directory with a git repo
	dir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	// Add a remote
	cmd = exec.Command("git", "remote", "add", "origin", "git@github.com:faraz/myproject.git")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git remote add: %v", err)
	}

	project, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("DetectProject: %v", err)
	}

	if project != "myproject" {
		t.Errorf("got %q, want %q", project, "myproject")
	}
}

func TestDetectProject_FallbackToDirectory(t *testing.T) {
	// Create a temp directory without git
	dir := t.TempDir()
	subdir := filepath.Join(dir, "my-cool-project")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	project, err := DetectProject(subdir)
	if err != nil {
		t.Fatalf("DetectProject: %v", err)
	}

	if project != "my-cool-project" {
		t.Errorf("got %q, want %q", project, "my-cool-project")
	}
}
