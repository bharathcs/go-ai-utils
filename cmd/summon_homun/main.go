package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	gitRoot, err := getGitRoot()
	if err != nil {
		fmt.Printf("%s\n", errorStyle.Render("Error: Not in a git repository"))
		os.Exit(1)
	}

	m := initialModel()
	m.gitRoot = gitRoot

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
