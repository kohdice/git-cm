package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// commitMessage is a struct that holds the fields for a commit message.
type commitMessage struct {
	Prefix      string
	Summary     string
	Description string
}

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

// checkStagedFiles checks if there are any staged files in the repository.
// If no staged files are found, it returns an error.
func checkStagedFiles(wt *git.Worktree) error {
	status, err := wt.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	for _, s := range status {
		if s.Staging != git.Unmodified {
			return nil
		}
	}

	return fmt.Errorf("no files are staged")
}

// commitRepo commits changes using the provided commit message and author information.
// It returns the commit hash or an error.
func commitRepo(r *git.Repository, a author, m *commitMessage) (string, error) {
	wt, err := r.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := checkStagedFiles(wt); err != nil {
		return "", err
	}

	msg := fmt.Sprintf("%s: %s\n\n%s", m.Prefix, m.Summary, m.Description)

	h, err := wt.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  a.Name,
			Email: a.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	return h.String(), nil
}
