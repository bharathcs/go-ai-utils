package main

import (
	"testing"
)

func TestExtractRepoFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
		wantErr  bool
	}{
		{
			name:     "GitHub HTTPS URL",
			url:      "https://github.com/owner/repo.git",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "GitHub HTTPS URL without .git",
			url:      "https://github.com/owner/repo",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "GitHub SSH URL",
			url:      "git@github.com:owner/repo.git",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "GitHub SSH URL without .git",
			url:      "git@github.com:owner/repo",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "Gitea HTTPS URL",
			url:      "https://gitea.example.com/owner/repo.git",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "Gitea SSH URL",
			url:      "git@gitea.example.com:owner/repo.git",
			expected: "owner/repo",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractRepoFromURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractRepoFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("extractRepoFromURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsGitHub(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "GitHub HTTPS URL",
			url:  "https://github.com/owner/repo.git",
			want: true,
		},
		{
			name: "GitHub SSH URL",
			url:  "git@github.com:owner/repo.git",
			want: true,
		},
		{
			name: "Gitea URL",
			url:  "https://gitea.example.com/owner/repo.git",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGitHub(tt.url); got != tt.want {
				t.Errorf("isGitHub() = %v, want %v", got, tt.want)
			}
		})
	}
}
