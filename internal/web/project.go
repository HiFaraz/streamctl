package web

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectProject infers the project name from git remote or directory name.
func DetectProject(dir string) (string, error) {
	// Try git remote first
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err == nil {
		return parseRepoName(strings.TrimSpace(string(out))), nil
	}

	// Fall back to directory name
	return filepath.Base(dir), nil
}

// parseRepoName extracts repo name from git remote URL.
// Handles: git@github.com:user/repo.git, https://github.com/user/repo.git
func parseRepoName(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format: git@github.com:user/repo
	if idx := strings.LastIndex(url, ":"); idx != -1 && !strings.Contains(url, "://") {
		url = url[idx+1:]
	}

	// Handle HTTPS format: https://github.com/user/repo
	if idx := strings.LastIndex(url, "/"); idx != -1 {
		url = url[idx+1:]
	}

	return url
}
