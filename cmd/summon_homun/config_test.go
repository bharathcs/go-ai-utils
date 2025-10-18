package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_WithFile(t *testing.T) {
	// Save original XDG_CONFIG_HOME
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalConfigHome)

	// Create temporary directory with config file
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	configDir := filepath.Join(tmpDir, "homun")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configContent := `github:
  username: testuser
  base_url: https://github.com

gitea:
  username: giteauser
  base_url: https://gitea.example.com
`

	configPath := filepath.Join(configDir, "config.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config.GitHub.Username != "testuser" {
		t.Errorf("GitHub.Username = %s, want testuser", config.GitHub.Username)
	}
	if config.GitHub.BaseURL != "https://github.com" {
		t.Errorf("GitHub.BaseURL = %s, want https://github.com", config.GitHub.BaseURL)
	}
	if config.Gitea.Username != "giteauser" {
		t.Errorf("Gitea.Username = %s, want giteauser", config.Gitea.Username)
	}
	if config.Gitea.BaseURL != "https://gitea.example.com" {
		t.Errorf("Gitea.BaseURL = %s, want https://gitea.example.com", config.Gitea.BaseURL)
	}
}

func TestGetRepoName(t *testing.T) {
	config := &Config{
		GitHub: GitHubConfig{
			Username: "bharathcs",
			BaseURL:  "https://github.com",
		},
		Gitea: GiteaConfig{
			Username: "bhcs",
			BaseURL:  "https://gitea.example.com",
		},
	}

	tests := []struct {
		name       string
		repoURL    string
		folderName string
		want       string
	}{
		{
			name:       "GitHub URL",
			repoURL:    "https://github.com/some/repo.git",
			folderName: "myproject",
			want:       "bharathcs/myproject",
		},
		{
			name:       "Gitea URL",
			repoURL:    "https://gitea.example.com/some/repo.git",
			folderName: "myproject",
			want:       "bhcs/myproject",
		},
		{
			name:       "Empty URL defaults to GitHub",
			repoURL:    "",
			folderName: "myproject",
			want:       "bharathcs/myproject",
		},
		{
			name:       "URL with gitea keyword",
			repoURL:    "https://my-gitea.com/some/repo.git",
			folderName: "myproject",
			want:       "bhcs/myproject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.GetRepoName(tt.repoURL, tt.folderName)
			if got != tt.want {
				t.Errorf("GetRepoName() = %s, want %s", got, tt.want)
			}
		})
	}
}
