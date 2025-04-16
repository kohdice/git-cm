package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"gopkg.in/ini.v1"
)

// author holds the name and email of a Git user.
type author struct {
	Name  string
	Email string
}

// loadGlobalConfig loads the global Git configuration from the user's ~/.gitconfig file.
// It returns an ini.File object representing the configuration, or an error if it fails.
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

// loadGlobalAuthor retrieves the user's name and email from the [user] section of the global
// Git configuration (stored in ~/.gitconfig). It returns an author pointer populated with the
// retrieved data, or an error if any required information is missing.
func loadGlobalAuthor() (*author, error) {
	cfg, err := loadGlobalConfig()
	if err != nil {
		return nil, err
	}

	userSection, err := cfg.GetSection("user")
	if err != nil {
		return nil, fmt.Errorf("failed to get [user] section: %w", err)
	}

	nameKey, err := userSection.GetKey("name")
	if err != nil {
		return nil, fmt.Errorf("failed to get user.name key: %w", err)
	}

	emailKey, err := userSection.GetKey("email")
	if err != nil {
		return nil, fmt.Errorf("failed to get user.email key: %w", err)
	}
	return &author{
		Name:  nameKey.String(),
		Email: emailKey.String(),
	}, nil
}

// getAuthorInfo obtains author information from the repository's configuration.
// If the repository configuration does not provide user information,
// the function falls back to loading the global Git configuration.
// It returns an author pointer containing the name and email, or an error.
func getAuthorInfo(repo *git.Repository) (*author, error) {
	cfg, err := repo.Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository config: %w", err)
	}

	if cfg.User.Name != "" && cfg.User.Email != "" {
		return &author{
			Name:  cfg.User.Name,
			Email: cfg.User.Email,
		}, nil
	}
	return loadGlobalAuthor()
}
