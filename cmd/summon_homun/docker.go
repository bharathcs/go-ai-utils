package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

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

type dockerOutputMsg struct {
	line   string
	closed bool
}

func waitForOutput(sub <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-sub
		if !ok {
			return dockerOutputMsg{closed: true}
		}
		return dockerOutputMsg{line: line, closed: false}
	}
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
			"-e", "HOMUNCULUS_GH_API_KEY",
			"-e", fmt.Sprintf("HOMUNCULUS_REPO_URL=%s", repoURL),
			"-e", fmt.Sprintf("HOMUNCULUS_REPO=%s", repo),
			"-e", fmt.Sprintf("HOMUNCULUS_BRANCH=%s", m.branch),
			"-e", "HOMUNCULUS_SSH_KEY_PRIVATE",
			"-e", "HOMUNCULUS_SSH_KEY_PUBLIC",
			"-v", fmt.Sprintf("%s:/workspace", workspacePath),
			"-v", fmt.Sprintf("%s:/report", reportPath),
			"-w", "/workspace",
			"homun-dev",
			"bash", "-c", "/home/homun/runner.bash",
		)

		cmd.Stdin = os.Stdin

		// Create pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return dockerRunMsg{err: fmt.Errorf("failed to create stdout pipe: %w", err)}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return dockerRunMsg{err: fmt.Errorf("failed to create stderr pipe: %w", err)}
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return dockerRunMsg{err: fmt.Errorf("failed to start docker: %w", err)}
		}

		// Create channel for output
		outputChan := make(chan string, 100)

		var wg sync.WaitGroup
		wg.Add(2)

		// Read from stdout
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				outputChan <- scanner.Text()
			}
		}()

		// Read from stderr
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				outputChan <- scanner.Text()
			}
		}()

		// Wait for command to complete and close channel
		go func() {
			wg.Wait()
			cmd.Wait()
			close(outputChan)
		}()

		// Store the channel in the global output map
		dockerOutputChannels[m.branch] = outputChan

		return dockerRunMsg{err: nil}
	}
}

// Global map to store output channels per branch
var (
	dockerOutputChannels = make(map[string]chan string)
	dockerOutputMutex    sync.RWMutex
)

func getOutputChannel(branch string) <-chan string {
	dockerOutputMutex.RLock()
	defer dockerOutputMutex.RUnlock()
	return dockerOutputChannels[branch]
}
