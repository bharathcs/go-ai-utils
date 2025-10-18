package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	StateInit State = iota
	StateValidating
	StateInputRepoURL
	StateInputRepo
	StateInputBranch
	StateEditInstructions
	StateConfirm
	StateRunning
	StateDone
	StateError
)

type validationMsg struct {
	err error
}

type dockerRunMsg struct {
	err error
}

type timerTickMsg time.Time

type model struct {
	state            State
	err              error
	repoURL          string
	repo             string
	branch           string
	gitRoot          string
	branchDir        string
	currentUser      string
	elapsed          time.Duration
	inputBuffer      string
	validationErr    string
	textarea         textarea.Model
	instructionsPath string
	dockerOutput     []string
	terminalHeight   int
	terminalWidth    int
	config           *Config
	issueContent     string
}

func initialModel() model {
	return model{
		state:       StateInit,
		inputBuffer: "",
	}
}

func (m model) Init() tea.Cmd {
	return validate()
}
