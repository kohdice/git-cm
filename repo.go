package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// findRepoRoot searches upward from the current working directory until it finds
// the root of a Git repository (i.e. a directory containing a ".git" folder).
func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("git repository not found")
		}

		dir = parent
	}
}

// openRepo opens the Git repository located at the specified directory.
// The provided directory is expected to be the root of a valid Git repository.
func openRepo(dir string) (*git.Repository, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository at %s: %w", dir, err)
	}
	return repo, nil
}
