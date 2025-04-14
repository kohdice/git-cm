package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestFindRepoRoot(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(origDir)
	}()

	tests := []struct {
		name    string
		prepare func(t *testing.T) (cleanup func(), expected string, expectErr bool)
	}{
		{
			name: "RepositoryAtCurrentDirectory",
			prepare: func(t *testing.T) (func(), string, bool) {
				tmpDir := t.TempDir()
				if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
					t.Fatalf("failed to create .git directory: %v", err)
				}

				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				cleanup := func() {
					_ = os.Chdir(origDir)
				}
				return cleanup, tmpDir, false
			},
		},
		{
			name: "RepositoryInParentDirectory",
			prepare: func(t *testing.T) (func(), string, bool) {
				tmpDir := t.TempDir()
				if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
					t.Fatalf("failed to create .git directory: %v", err)
				}

				subDir := filepath.Join(tmpDir, "sub")
				if err := os.Mkdir(subDir, 0755); err != nil {
					t.Fatalf("failed to create subdirectory: %v", err)
				}

				if err := os.Chdir(subDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				cleanup := func() {
					_ = os.Chdir(origDir)
				}
				return cleanup, tmpDir, false
			},
		},
		{
			name: "RepositoryNotFound",
			prepare: func(t *testing.T) (func(), string, bool) {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				cleanup := func() {
					_ = os.Chdir(origDir)
				}
				return cleanup, "", true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, expected, expectErr := tt.prepare(t)
			defer cleanup()

			root, err := findRepoRoot()
			if expectErr {
				if err == nil {
					t.Error("expected error, but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			canonicalExpected, err := filepath.EvalSymlinks(expected)
			if err != nil {
				t.Fatalf("failed to evaluate expected symlinks: %v", err)
			}

			canonicalRoot, err := filepath.EvalSymlinks(root)
			if err != nil {
				t.Fatalf("failed to evaluate root symlinks: %v", err)
			}

			if canonicalExpected != canonicalRoot {
				t.Errorf("expected root %q, but got %q", canonicalExpected, canonicalRoot)
			}
		})
	}
}

func TestOpenRepo(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (repoDir string, cleanup func())
		wantErr bool
	}{
		{
			name: "ValidGitRepository",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				_, err := git.PlainInit(tmpDir, false)
				if err != nil {
					t.Fatalf("failed to initialize git repository: %v", err)
				}
				return tmpDir, func() {}
			},
			wantErr: false,
		},
		{
			name: "InvalidGitRepository",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			wantErr: true,
		},
		{
			name: "NonExistentDirectory",
			setup: func(t *testing.T) (string, func()) {
				nonexistent := filepath.Join(os.TempDir(), "nonexistent-repo-directory")
				return nonexistent, func() {}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup := tt.setup(t)
			defer cleanup()

			repo, err := openRepo(dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("openRepo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if repo == nil && !tt.wantErr {
				t.Errorf("expected non-nil repository for valid directory")
			}
		})
	}
}

func TestCheckStagedFiles(t *testing.T) {
	t.Run("Has staged files", func(t *testing.T) {
		repoDir := t.TempDir()

		repo, err := git.PlainInit(repoDir, false)
		if err != nil {
			t.Fatalf("failed to init repo: %v", err)
		}

		dummyFile := filepath.Join(repoDir, "file.txt")
		if err := os.WriteFile(dummyFile, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		wt, err := repo.Worktree()
		if err != nil {
			t.Fatalf("failed to get worktree: %v", err)
		}

		if _, err := wt.Add("file.txt"); err != nil {
			t.Fatalf("failed to stage file: %v", err)
		}

		if err := checkStagedFiles(wt); err != nil {
			t.Errorf("expected no error for staged file, got: %v", err)
		}
	})

	t.Run("No staged files", func(t *testing.T) {
		repoDir := t.TempDir()

		repo, err := git.PlainInit(repoDir, false)
		if err != nil {
			t.Fatalf("failed to init repo: %v", err)
		}

		wt, err := repo.Worktree()
		if err != nil {
			t.Fatalf("failed to get worktree: %v", err)
		}

		if err := checkStagedFiles(wt); err == nil {
			t.Error("expected error for no staged files, got nil")
		}
	})
}

func TestCommitRepo(t *testing.T) {
	tests := []struct {
		name           string
		msg            commitMessage
		setupRepo      func(repoDir string, repo *git.Repository) error
		expectErr      bool
		expectedSubstr string
	}{
		{
			name: "Successful commit",
			msg: commitMessage{
				Prefix:      "feat",
				Summary:     "A new feature",
				Description: "Detailed description of the feature.",
			},
			setupRepo: func(repoDir string, repo *git.Repository) error {
				dummyFile := filepath.Join(repoDir, "dummy.txt")
				if err := os.WriteFile(dummyFile, []byte("dummy content"), 0644); err != nil {
					return fmt.Errorf("failed to write dummy file: %w", err)
				}
				wt, err := repo.Worktree()
				if err != nil {
					return fmt.Errorf("failed to get worktree: %w", err)
				}
				_, err = wt.Add("dummy.txt")
				return err
			},
			expectErr:      false,
			expectedSubstr: "feat: A new feature",
		},
		{
			name: "Worktree error",
			msg: commitMessage{
				Prefix:      "fix",
				Summary:     "A bugfix",
				Description: "Fixing a critical bug.",
			},
			setupRepo: func(repoDir string, repo *git.Repository) error {
				gitPath := filepath.Join(repoDir, ".git")
				return os.RemoveAll(gitPath)
			},
			expectErr:      true,
			expectedSubstr: "",
		},
		{
			name: "Empty commit error",
			msg: commitMessage{
				Prefix:      "docs",
				Summary:     "Documentation update",
				Description: "No changes made.",
			},
			setupRepo:      nil,
			expectErr:      true,
			expectedSubstr: "",
		},
	}

	author := author{
		Name:  "Tester",
		Email: "tester@example.com",
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repoDir, err := os.MkdirTemp("", "testrepo")
			if err != nil {
				t.Fatalf("failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(repoDir) // nolint:errcheck

			repo, err := git.PlainInit(repoDir, false)
			if err != nil {
				t.Fatalf("failed to initialize git repository: %v", err)
			}

			if tc.setupRepo != nil {
				if err := tc.setupRepo(repoDir, repo); err != nil {
					t.Fatalf("setupRepo failed: %v", err)
				}
			}

			commitHash, err := commitRepo(repo, author, &tc.msg)
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error, but got none; commitHash=%q", commitHash)
				} else if !strings.Contains(err.Error(), "failed") &&
					!strings.Contains(err.Error(), "empty commit") &&
					!strings.Contains(err.Error(), "no files are staged") {
					t.Errorf("expected an error containing 'failed', 'empty commit', or 'no files are staged', but got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if commitHash == "" {
				t.Error("expected non-empty commit hash")
			}

			time.Sleep(100 * time.Millisecond)

			commits, err := repo.Log(&git.LogOptions{})
			if err != nil {
				t.Fatalf("failed to get commit log: %v", err)
			}
			defer commits.Close()

			found := false
			err = commits.ForEach(func(c *object.Commit) error {
				if strings.Contains(c.Message, tc.expectedSubstr) {
					found = true
				}
				return nil
			})
			if err != nil {
				t.Fatalf("error iterating commits: %v", err)
			}

			if !found && tc.expectedSubstr != "" {
				t.Errorf("expected commit log to contain %q, but it did not", tc.expectedSubstr)
			}
		})
	}
}
