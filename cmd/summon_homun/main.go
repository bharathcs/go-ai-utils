package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command-line flags
	issueNumber := flag.String("issue", "", "Issue number to fetch and pre-fill in instructions")
	flag.Parse()

	gitRoot, err := getGitRoot()
	if err != nil {
		fmt.Printf("%s\n", errorStyle.Render("Error: Not in a git repository"))
		os.Exit(1)
	}

	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("%s\n", errorStyle.Render(fmt.Sprintf("Error loading config: %v", err)))
		os.Exit(1)
	}

	// Fetch issue content if --issue flag is provided
	var issueContent string
	if *issueNumber != "" {
		issueContent, err = fetchIssue(*issueNumber, gitRoot, config)
		if err != nil {
			fmt.Printf("%s\n", errorStyle.Render(fmt.Sprintf("Error fetching issue: %v", err)))
			os.Exit(1)
		}
	}

	m := initialModel()
	m.gitRoot = gitRoot
	m.config = config
	m.issueContent = issueContent

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
