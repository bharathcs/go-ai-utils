package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the configuration for Homunculus
type Config struct {
	GitHub GitHubConfig `yaml:"github"`
	Gitea  GiteaConfig  `yaml:"gitea"`
}

// GitHubConfig holds GitHub-specific configuration
type GitHubConfig struct {
	Username string `yaml:"username"`
	BaseURL  string `yaml:"base_url"`
}

// GiteaConfig holds Gitea-specific configuration
type GiteaConfig struct {
	Username string `yaml:"username"`
	BaseURL  string `yaml:"base_url"`
}

// LoadConfig loads configuration from ~/.config/homun/config.yml
// If the file doesn't exist, it returns default configuration
func LoadConfig() (*Config, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	configPath := filepath.Join(currentUser.HomeDir, ".config", "homun", "config.yml")

	// If config file doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultConfig(currentUser.Username), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for any missing values
	applyDefaults(&config, currentUser.Username)

	return &config, nil
}

// getDefaultConfig returns the default configuration
func getDefaultConfig(systemUsername string) *Config {
	return &Config{
		GitHub: GitHubConfig{
			Username: systemUsername,
			BaseURL:  "https://github.com",
		},
		Gitea: GiteaConfig{
			Username: systemUsername,
			BaseURL:  "https://gitea.com",
		},
	}
}

// applyDefaults fills in any missing values with defaults
func applyDefaults(config *Config, systemUsername string) {
	if config.GitHub.Username == "" {
		config.GitHub.Username = systemUsername
	}
	if config.GitHub.BaseURL == "" {
		config.GitHub.BaseURL = "https://github.com"
	}
	if config.Gitea.Username == "" {
		config.Gitea.Username = systemUsername
	}
	if config.Gitea.BaseURL == "" {
		config.Gitea.BaseURL = "https://gitea.com"
	}
}

// GetRepoName returns the appropriate repository name based on the remote URL
func (c *Config) GetRepoName(repoURL, folderName string) string {
	// Check if URL contains github
	if contains(repoURL, "github.com") {
		return fmt.Sprintf("%s/%s", c.GitHub.Username, folderName)
	}

	// Check if URL contains gitea or matches gitea base URL
	if contains(repoURL, "gitea") || contains(repoURL, c.Gitea.BaseURL) {
		return fmt.Sprintf("%s/%s", c.Gitea.Username, folderName)
	}

	// Default to GitHub username
	return fmt.Sprintf("%s/%s", c.GitHub.Username, folderName)
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(len(s) >= len(substr)) &&
		(s == substr || len(s) > len(substr) && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
