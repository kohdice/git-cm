package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
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
