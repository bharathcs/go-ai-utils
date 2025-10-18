package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// fetchIssue fetches the issue content from GitHub or Gitea using gh or tea CLI
func fetchIssue(issueNumber, gitRoot string, config *Config) (string, error) {
	// Get the remote URL to determine the platform
	remoteURL, err := getGitRemoteURL(gitRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get git remote URL: %w", err)
	}

	// Determine repository owner/name from remote URL
	repo, err := extractRepoFromURL(remoteURL)
	if err != nil {
		return "", fmt.Errorf("failed to extract repository from URL: %w", err)
	}

	// Detect platform and fetch issue
	if isGitHub(remoteURL) {
		return fetchGitHubIssue(issueNumber, repo)
	} else if isGitea(remoteURL, config) {
		return fetchGiteaIssue(issueNumber, repo)
	}

	return "", fmt.Errorf("unsupported git hosting platform")
}

// isGitHub checks if the URL is a GitHub URL
func isGitHub(url string) bool {
	return strings.Contains(url, "github.com")
}

// isGitea checks if the URL is a Gitea URL
func isGitea(url string, config *Config) bool {
	return strings.Contains(url, "gitea") || strings.Contains(url, config.Gitea.BaseURL)
}

// extractRepoFromURL extracts the owner/repo format from a git remote URL
// Examples:
//   - https://github.com/owner/repo.git -> owner/repo
//   - git@github.com:owner/repo.git -> owner/repo
//   - https://gitea.example.com/owner/repo.git -> owner/repo
func extractRepoFromURL(url string) (string, error) {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format (git@host:owner/repo)
	if strings.HasPrefix(url, "git@") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			return parts[1], nil
		}
	}

	// Handle HTTPS format (https://host/owner/repo)
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			// Get the last two parts (owner/repo)
			return strings.Join(parts[len(parts)-2:], "/"), nil
		}
	}

	return "", fmt.Errorf("unable to parse repository from URL: %s", url)
}

// fetchGitHubIssue fetches an issue from GitHub using gh CLI
func fetchGitHubIssue(issueNumber, repo string) (string, error) {
	// Check if gh is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return "", fmt.Errorf("gh CLI is not installed. Please install it from https://cli.github.com/")
	}

	// Run gh issue view command with --comments flag to get full issue details
	cmd := exec.Command("gh", "issue", "view", issueNumber, "--repo", repo, "--comments")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to fetch GitHub issue: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// fetchGiteaIssue fetches an issue from Gitea using tea CLI
func fetchGiteaIssue(issueNumber, repo string) (string, error) {
	// Check if tea is installed
	if _, err := exec.LookPath("tea"); err != nil {
		return "", fmt.Errorf("tea CLI is not installed. Please install it from https://gitea.com/gitea/tea")
	}

	// Run tea issue view command
	cmd := exec.Command("tea", "issue", "view", issueNumber, "--repo", repo)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to fetch Gitea issue: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// getGitRemoteURL gets the remote URL for the git repository
func getGitRemoteURL(gitRoot string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "remote", "get-url", "origin")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
