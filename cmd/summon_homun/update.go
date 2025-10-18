package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.state == StateEditInstructions {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				if msg.Type == tea.KeyEsc && m.textarea.Focused() {
					m.textarea.Blur()
					return m, nil
				}
				return m, tea.Quit
			case tea.KeyCtrlS:
				if err := m.saveInstructions(); err != nil {
					m.state = StateError
					m.err = err
					return m, nil
				}
				m.state = StateConfirm
				return m, nil
			default:
				m.textarea, cmd = m.textarea.Update(msg)
				return m, cmd
			}
		default:
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalHeight = msg.Height
		m.terminalWidth = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			return m.handleEnter()

		case "backspace":
			if len(m.inputBuffer) > 0 {
				m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
			}

		default:
			if m.state == StateInputRepoURL || m.state == StateInputRepo || m.state == StateInputBranch {
				m.inputBuffer += msg.String()
			} else if m.state == StateConfirm {
				switch strings.ToLower(msg.String()) {
				case "y":
					m.state = StateRunning
					return m, tea.Batch(runDocker(m), tickTimer())
				case "n":
					return m, tea.Quit
				}
			}
		}

	case validationMsg:
		if msg.err != nil {
			m.state = StateError
			m.err = msg.err
			return m, nil
		}
		return m.advanceFromValidation()

	case dockerRunMsg:
		if msg.err != nil {
			m.state = StateDone
			m.err = msg.err
			return m, tea.Quit
		}
		// Docker started, now wait for output
		outputChan := getOutputChannel(m.branch)
		if outputChan != nil {
			return m, waitForOutput(outputChan)
		}
		return m, nil

	case dockerOutputMsg:
		if msg.closed {
			// Docker process finished
			m.state = StateDone
			return m, tea.Quit
		}
		if msg.line != "" {
			m.dockerOutput = append(m.dockerOutput, msg.line)
			// Keep only the last 1000 lines to avoid memory issues
			if len(m.dockerOutput) > 1000 {
				m.dockerOutput = m.dockerOutput[len(m.dockerOutput)-1000:]
			}
		}
		// Continue listening for more output
		outputChan := getOutputChannel(m.branch)
		if outputChan != nil {
			return m, waitForOutput(outputChan)
		}
		return m, nil

	case timerTickMsg:
		if m.state == StateRunning {
			m.elapsed += time.Second
			return m, tickTimer()
		}
	}

	return m, nil
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case StateInputRepoURL:
		m.repoURL = strings.TrimSpace(m.inputBuffer)
		if m.repoURL == "" {
			m.repoURL = getDefaultRepoURL()
		}
		m.inputBuffer = ""
		m.state = StateInputRepo
		return m, nil

	case StateInputRepo:
		m.repo = strings.TrimSpace(m.inputBuffer)
		if m.repo == "" {
			m.repo = getDefaultRepo(m.config)
		}
		m.inputBuffer = ""
		m.state = StateInputBranch
		return m, nil

	case StateInputBranch:
		branch := strings.TrimSpace(m.inputBuffer)
		if !isValidBranchName(branch) {
			m.validationErr = "Branch name must contain only alphanumeric characters, hyphens, or underscores"
			m.inputBuffer = ""
			return m, nil
		}
		m.branch = branch
		m.validationErr = ""
		return m.setupDirectories()
	}

	return m, nil
}

func (m model) advanceFromValidation() (tea.Model, tea.Cmd) {
	repoURLEnv := os.Getenv("HOMUNCULUS_REPO_URL")
	repoEnv := os.Getenv("HOMUNCULUS_REPO")

	if repoURLEnv == "" {
		m.state = StateInputRepoURL
		return m, nil
	}
	m.repoURL = repoURLEnv

	if repoEnv == "" {
		m.state = StateInputRepo
		return m, nil
	}
	m.repo = repoEnv

	m.state = StateInputBranch
	return m, nil
}

func (m model) setupDirectories() (tea.Model, tea.Cmd) {
	branchDir := filepath.Join(m.gitRoot, ".homun", "branches", m.branch)
	m.branchDir = branchDir

	if err := os.MkdirAll(filepath.Join(branchDir, "report"), 0775); err != nil {
		m.state = StateError
		m.err = fmt.Errorf("failed to create report directory: %w", err)
		return m, nil
	}

	if err := os.MkdirAll(filepath.Join(branchDir, "workspace"), 0775); err != nil {
		m.state = StateError
		m.err = fmt.Errorf("failed to create workspace directory: %w", err)
		return m, nil
	}

	currentUser, err := user.Current()
	if err != nil {
		m.state = StateError
		m.err = fmt.Errorf("failed to get current user: %w", err)
		return m, nil
	}
	m.currentUser = currentUser.Username

	chownCmd := exec.Command("chown", "-R", fmt.Sprintf("%s:homun", m.currentUser), branchDir)
	if err := chownCmd.Run(); err != nil {
		m.state = StateError
		m.err = fmt.Errorf("failed to set ownership (you may need sudo): %w", err)
		return m, nil
	}

	chmodCmd := exec.Command("chmod", "-R", "g=u", branchDir)
	if err := chmodCmd.Run(); err != nil {
		m.state = StateError
		m.err = fmt.Errorf("failed to set permissions: %w", err)
		return m, nil
	}

	m.instructionsPath = filepath.Join(branchDir, "report", "instructions.md")

	ta := textarea.New()
	ta.Placeholder = "Enter your instructions for Homunculus..."
	ta.SetWidth(80)
	ta.SetHeight(20)
	ta.Focus()
	m.textarea = ta

	if err := m.loadInstructions(); err != nil {
		m.state = StateError
		m.err = err
		return m, nil
	}

	m.state = StateEditInstructions
	m.inputBuffer = ""
	return m, textarea.Blink
}

func tickTimer() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return timerTickMsg(t)
	})
}
