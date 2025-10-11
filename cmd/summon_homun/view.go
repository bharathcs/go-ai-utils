package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(asciiArt))
	b.WriteString("\n")
	b.WriteString(taglineStyle.Render(tagline))
	b.WriteString("\n\n")

	switch m.state {
	case StateInit, StateValidating:
		b.WriteString(infoStyle.Render("Validating environment..."))

	case StateInputRepoURL:
		b.WriteString(promptStyle.Render("HOMUNCULUS_REPO_URL not set."))
		b.WriteString("\n")
		defaultURL := getDefaultRepoURL()
		if defaultURL != "" {
			b.WriteString(infoStyle.Render(fmt.Sprintf("Default: %s", defaultURL)))
			b.WriteString("\n")
		}
		b.WriteString(promptStyle.Render("Enter repository URL (or press Enter for default): "))
		b.WriteString(m.inputBuffer)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("█"))

	case StateInputRepo:
		b.WriteString(promptStyle.Render("HOMUNCULUS_REPO not set."))
		b.WriteString("\n")
		defaultRepo := getDefaultRepo()
		if defaultRepo != "" {
			b.WriteString(infoStyle.Render(fmt.Sprintf("Default: %s", defaultRepo)))
			b.WriteString("\n")
		}
		b.WriteString(promptStyle.Render("Enter repository name (or press Enter for default): "))
		b.WriteString(m.inputBuffer)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("█"))

	case StateInputBranch:
		if m.validationErr != "" {
			b.WriteString(errorStyle.Render("Error: " + m.validationErr))
			b.WriteString("\n")
		}
		b.WriteString(promptStyle.Render("Enter branch name (alphanumeric, hyphen, underscore only): "))
		b.WriteString(m.inputBuffer)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("█"))

	case StateEditInstructions:
		b.WriteString(successStyle.Render("Edit Instructions"))
		b.WriteString("\n\n")
		b.WriteString(infoStyle.Render(fmt.Sprintf("Editing: %s", m.instructionsPath)))
		b.WriteString("\n\n")
		b.WriteString(m.textarea.View())
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Ctrl+S to save and continue | Ctrl+C to quit"))

	case StateConfirm:
		b.WriteString(successStyle.Render("Environment validated!"))
		b.WriteString("\n\n")
		b.WriteString(infoStyle.Render("Configuration:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Repository URL: %s\n", m.repoURL))
		b.WriteString(fmt.Sprintf("  Repository: %s\n", m.repo))
		b.WriteString(fmt.Sprintf("  Branch: %s\n", m.branch))
		b.WriteString(fmt.Sprintf("  Branch Directory: %s\n", m.branchDir))
		b.WriteString("\n")
		b.WriteString(warningStyle.Render("Ready to summon Homunculus. Continue? (y/n): "))

	case StateRunning:
		b.WriteString(successStyle.Render("Running Homunculus..."))
		b.WriteString("\n\n")
		b.WriteString(infoStyle.Render(fmt.Sprintf("Elapsed time: %d seconds", int(m.elapsed.Seconds()))))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Ctrl+C to cancel"))

	case StateDone:
		if m.err != nil {
			b.WriteString(errorStyle.Render("Homunculus exited with error:"))
			b.WriteString("\n")
			b.WriteString(errorStyle.Render(m.err.Error()))
		} else {
			b.WriteString(successStyle.Render("Homunculus completed successfully!"))
		}

	case StateError:
		b.WriteString(errorStyle.Render("Error:"))
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
	}

	b.WriteString("\n")
	return b.String()
}
