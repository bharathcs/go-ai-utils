package main

import (
	"fmt"
	"os"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"
)

func validate() tea.Cmd {
	return func() tea.Msg {
		imageExists, err := checkDockerImage("homun-dev")
		if err != nil {
			return validationMsg{err: fmt.Errorf("failed to check Docker image: %w", err)}
		}
		if !imageExists {
			return validationMsg{err: fmt.Errorf("Docker image 'homun-dev' not found. Please build it first")}
		}

		requiredEnvVars := []string{
			"ANTHROPIC_API_KEY",
			"HOMUNCULUS_SSH_KEY_PRIVATE",
			"HOMUNCULUS_SSH_KEY_PUBLIC",
			"HOMUNCULUS_GITEA_API_KEY",
		}

		for _, envVar := range requiredEnvVars {
			if os.Getenv(envVar) == "" {
				return validationMsg{err: fmt.Errorf("required environment variable %s is not set", envVar)}
			}
		}

		return validationMsg{err: nil}
	}
}

func isValidBranchName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}
