package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func checkDockerImage(imageName string) (bool, error) {
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}", "--filter", fmt.Sprintf("reference=%s", imageName))
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == imageName {
			return true, nil
		}
	}
	return false, nil
}

func runDocker(m model) tea.Cmd {
	return func() tea.Msg {
		gitRoot, err := getGitRoot()
		if err != nil {
			return dockerRunMsg{err: fmt.Errorf("failed to get git root: %w", err)}
		}

		workspacePath := filepath.Join(gitRoot, ".homun", "branches", m.branch, "workspace")
		reportPath := filepath.Join(gitRoot, ".homun", "branches", m.branch, "report")

		repoURL := m.repoURL
		if repoURL == "" {
			repoURL = getDefaultRepoURL()
		}

		repo := m.repo
		if repo == "" {
			repo = getDefaultRepo()
		}

		cmd := exec.Command("docker", "run", "-it", "--rm",
			"--cap-add=NET_ADMIN",
			"--cap-add=NET_RAW",
			"-e", "ANTHROPIC_API_KEY",
			"-e", "HOMUNCULUS_GITEA_API_KEY",
			"-e", fmt.Sprintf("HOMUNCULUS_REPO_URL=%s", repoURL),
			"-e", fmt.Sprintf("HOMUNCULUS_REPO=%s", repo),
			"-e", "HOMUNCULUS_SSH_KEY_PRIVATE",
			"-e", "HOMUNCULUS_SSH_KEY_PUBLIC",
			"-v", fmt.Sprintf("%s:/workspace", workspacePath),
			"-v", fmt.Sprintf("%s:/report", reportPath),
			"-w", "/workspace",
			"homun-dev",
			"bash", "-c", "/home/homun/runner.bash",
		)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return dockerRunMsg{err: err}
		}

		return dockerRunMsg{err: nil}
	}
}
