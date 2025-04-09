package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

func TestLoadGlobalConfig(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		expectErr     bool
	}{
		{
			name:          "ValidConfiguration",
			configContent: "[user]\nname = GlobalUser\nemail = global@example.com\n",
			expectErr:     false,
		},
		{
			name:          "EmptyConfigurationFile",
			configContent: "",
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			t.Setenv("HOME", tempDir)

			cfgPath := filepath.Join(tempDir, ".gitconfig")
			if err := os.WriteFile(cfgPath, []byte(tt.configContent), 0644); err != nil {
				t.Fatalf("failed to write .gitconfig: %v", err)
			}

			cfg, err := loadGlobalConfig()
			if tt.expectErr {
				if err == nil {
					t.Error("expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if cfg == nil {
					t.Error("expected non-nil config")
				}
			}
		})
	}
}

func TestLoadGlobalAuthor(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		expectedName  string
		expectedEmail string
		expectError   bool
	}{
		{
			name:          "ValidConfig",
			configContent: "[user]\nname = GlobalUser\nemail = global@example.com\n",
			expectedName:  "GlobalUser",
			expectedEmail: "global@example.com",
			expectError:   false,
		},
		{
			name:          "MissingUserSection",
			configContent: "[other]\nkey = value\n",
			expectError:   true,
		},
		{
			name:          "MissingNameKey",
			configContent: "[user]\nemail = global@example.com\n",
			expectError:   true,
		},
		{
			name:          "MissingEmailKey",
			configContent: "[user]\nname = GlobalUser\n",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			t.Setenv("HOME", tempDir)

			cfgPath := filepath.Join(tempDir, ".gitconfig")
			if err := os.WriteFile(cfgPath, []byte(tt.configContent), 0644); err != nil {
				t.Fatalf("failed to write .gitconfig: %v", err)
			}

			name, email, err := loadGlobalAuthor()
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if name != tt.expectedName {
					t.Errorf("expected name %q, got %q", tt.expectedName, name)
				}

				if email != tt.expectedEmail {
					t.Errorf("expected email %q, got %q", tt.expectedEmail, email)
				}
			}
		})
	}
}

func TestGetAuthorInfo(t *testing.T) {
	tests := []struct {
		name                string
		repoSetup           func(t *testing.T, repoDir string)
		globalConfigContent string
		expectedName        string
		expectedEmail       string
		expectError         bool
	}{
		{
			name: "RepositoryHasAuthorInfo",
			repoSetup: func(t *testing.T, repoDir string) {
				repo, err := git.PlainOpen(repoDir)
				if err != nil {
					t.Fatalf("failed to open repository: %v", err)
				}

				cfg, err := repo.Config()
				if err != nil {
					t.Fatalf("failed to get repository config: %v", err)
				}

				cfg.User.Name = "RepoUser"
				cfg.User.Email = "repo@example.com"
				if err := repo.Storer.SetConfig(cfg); err != nil {
					t.Fatalf("failed to set repository config: %v", err)
				}
			},
			globalConfigContent: "[user]\nname = GlobalUser\nemail = global@example.com\n",
			expectedName:        "RepoUser",
			expectedEmail:       "repo@example.com",
			expectError:         false,
		},
		{
			name: "FallbackToGlobal",
			repoSetup: func(t *testing.T, repoDir string) {
			},
			globalConfigContent: "[user]\nname = GlobalUser\nemail = global@example.com\n",
			expectedName:        "GlobalUser",
			expectedEmail:       "global@example.com",
			expectError:         false,
		},
		{
			name: "FallbackFailsDueToInvalidGlobalConfig",
			repoSetup: func(t *testing.T, repoDir string) {
			},
			globalConfigContent: "",
			expectError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoDir := t.TempDir()
			repo, err := git.PlainInit(repoDir, false)
			if err != nil {
				t.Fatalf("failed to initialize repository: %v", err)
			}

			globalHome := t.TempDir()
			t.Setenv("HOME", globalHome)

			cfgPath := filepath.Join(globalHome, ".gitconfig")
			if err := os.WriteFile(cfgPath, []byte(tt.globalConfigContent), 0644); err != nil {
				t.Fatalf("failed to write global .gitconfig: %v", err)
			}

			if tt.repoSetup != nil {
				tt.repoSetup(t, repoDir)
			}

			name, email, err := getAuthorInfo(repo)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if name != tt.expectedName {
					t.Errorf("expected name %q, got %q", tt.expectedName, name)
				}

				if email != tt.expectedEmail {
					t.Errorf("expected email %q, got %q", tt.expectedEmail, email)
				}
			}
		})
	}
}
