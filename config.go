package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"gopkg.in/ini.v1"
)

// loadGlobalConfig loads the global git configuration from ~/.gitconfig.
func loadGlobalConfig() (*ini.File, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cfgPath := filepath.Join(homeDir, ".gitconfig")
	cfg, err := ini.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load global config file: %w", err)
	}
	return cfg, nil
}

// loadGlobalAuthor retrieves the user's name and email from the [user] section of the global git config.
func loadGlobalAuthor() (string, string, error) {
	cfg, err := loadGlobalConfig()
	if err != nil {
		return "", "", err
	}

	userSection, err := cfg.GetSection("user")
	if err != nil {
		return "", "", fmt.Errorf("failed to get [user] section: %w", err)
	}

	nameKey, err := userSection.GetKey("name")
	if err != nil {
		return "", "", fmt.Errorf("failed to get user.name key: %w", err)
	}

	emailKey, err := userSection.GetKey("email")
	if err != nil {
		return "", "", fmt.Errorf("failed to get user.email key: %w", err)
	}
	return nameKey.String(), emailKey.String(), nil
}

// getAuthorInfo obtains author information from the repository config.
// If the repository doesn't have the user configuration, it falls back to the global git config.
func getAuthorInfo(repo *git.Repository) (string, string, error) {
	cfg, err := repo.Config()
	if err != nil {
		return "", "", fmt.Errorf("failed to get repository config: %w", err)
	}

	if cfg.User.Name != "" && cfg.User.Email != "" {
		return cfg.User.Name, cfg.User.Email, nil
	}
	return loadGlobalAuthor()
}
