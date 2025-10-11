package main

import "github.com/charmbracelet/lipgloss"

const asciiArt = `


 __    __
/  |  /  |
$$ |  $$ |  ______   _____  ____   __    __  _______
$$ |__$$ | /      \ /     \/    \ /  |  /  |/       \
$$    $$ |/$$$$$$  |$$$$$$ $$$$  |$$ |  $$ |$$$$$$$  |
$$$$$$$$ |$$ |  $$ |$$ | $$ | $$ |$$ |  $$ |$$ |  $$ |
$$ |  $$ |$$ \__$$ |$$ | $$ | $$ |$$ \__$$ |$$ |  $$ |
$$ |  $$ |$$    $$/ $$ | $$ | $$ |$$    $$/ $$ |  $$ |
$$/   $$/  $$$$$$/  $$/  $$/  $$/  $$$$$$/  $$/   $$/

`

const tagline = "bhcs/homunculus -- Autonomous Claude Code boxed up and tightly strapped"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	taglineStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42"))

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214"))
)
