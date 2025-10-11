package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

type State int

const (
	StateInit State = iota
	StateValidating
	StateInputRepoURL
	StateInputRepo
	StateInputBranch
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
	state         State
	err           error
	repoURL       string
	repo          string
	branch        string
	gitRoot       string
	branchDir     string
	currentUser   string
	elapsed       time.Duration
	inputBuffer   string
	validationErr string
}

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

func initialModel() model {
	return model{
		state:       StateInit,
		inputBuffer: "",
	}
}

func (m model) Init() tea.Cmd {
	return validate()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		m.state = StateDone
		if msg.err != nil {
			m.err = msg.err
		}
		return m, tea.Quit

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
			m.repo = getDefaultRepo()
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

	m.state = StateConfirm
	m.inputBuffer = ""
	return m, nil
}

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
			b.WriteString(errorStyle.Render("Error: "+m.validationErr))
			b.WriteString("\n")
		}
		b.WriteString(promptStyle.Render("Enter branch name (alphanumeric, hyphen, underscore only): "))
		b.WriteString(m.inputBuffer)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("█"))

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

func getGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getDefaultRepoURL() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getDefaultRepo() string {
	currentUser, err := user.Current()
	if err != nil {
		return ""
	}

	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	folderName := filepath.Base(cwd)
	return fmt.Sprintf("%s/%s", currentUser.Username, folderName)
}

func isValidBranchName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

func tickTimer() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return timerTickMsg(t)
	})
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
