package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	ai "github.com/bharathcs/go-ai-utils/lib"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/openai/openai-go"
)

type State int

const (
	StateInput State = iota
	StateLoading
	StateShowingSolutions
)

type CommandSolution struct {
	Command   string `json:"command"`
	Relevance int    `json:"relevance"` // 1-3, where 3 is most relevant
}

type CommandSolutions struct {
	Solutions []CommandSolution `json:"solutions"`
}

type apiResponseMsg struct {
	solutions []CommandSolution
}

type apiErrorMsg struct {
	err error
}

type model struct {
	state            State
	input            string
	cursor           int
	history          []string // Previous prompts and responses
	solutions        []CommandSolution
	selectedSolution int
	loadingFrame     int
	conversation     *ai.Conversation
	aiClient         *openai.Client
	aiConfig         *ai.Config
	pipedContext     string // Context from piped stdin
}

var (
	// Styles
	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	historyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	solutionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			PaddingLeft(2)

	selectedSolutionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Bold(true)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

func initialModel(pipedContext string) model {
	client, config, err := ai.NewClientFromEnv()
	var conversation *ai.Conversation
	var history []string

	if err == nil {
		systemPrompt := `You are a helpful command-line assistant. When users ask for help with commands, provide up to 3 specific, working command snippets. Each solution should only include the exact command and a relevance rating (3=most relevant, 1=least relevant). Focus on practical, commonly-used commands that can be executed immediately. Provide fewer solutions if 1-2 commands are sufficient.`
		conversation = ai.NewConversation(client, config, systemPrompt)
	} else {
		// Add warning to history if client couldn't be created
		history = []string{
			errorStyle.Render("⚠ Warning: AI client not configured"),
			historyStyle.Render("Set OPENAI_API_KEY environment variable to use AI features"),
			historyStyle.Render("You can still type queries, but they won't be processed."),
			"",
		}
	}

	if pipedContext != "" {
		contextPreview := pipedContext
		if len(contextPreview) > 50 {
			contextPreview = contextPreview[:50] + "..."
		}
		history = append(history,
			lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(":clipboard: Context from stdin: ")+
				historyStyle.Render(contextPreview),
			"",
		)
	}

	return model{
		state:            StateInput,
		input:            "",
		cursor:           0,
		history:          history,
		solutions:        []CommandSolution{},
		selectedSolution: -1,
		loadingFrame:     0,
		conversation:     conversation,
		aiClient:         client,
		aiConfig:         config,
		pipedContext:     pipedContext,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if m.state == StateInput {
				if strings.TrimSpace(m.input) != "" {
					// Record the prompt in history
					m.history = append(m.history, "> "+m.input)
					prompt := m.input

					// Add piped context if available
					if m.pipedContext != "" {
						prompt = fmt.Sprintf("Context:\n```\n%s\n```\n\nQuery: %s", m.pipedContext, prompt)
					}

					m.input = ""
					m.cursor = 0
					m.state = StateLoading
					m.selectedSolution = -1
					return m, tea.Batch(callAPI(prompt, m.aiClient, m.aiConfig), tickLoading())
				}
			} else if m.state == StateShowingSolutions && m.selectedSolution >= 0 {
				// User selected a solution - execute it (type it out)
				selected := m.solutions[m.selectedSolution].Command
				return m, tea.Sequence(
					executeSolution(selected),
					tea.Quit,
				)
			}

		case "up":
			if m.state == StateShowingSolutions {
				if m.selectedSolution == -1 {
					m.selectedSolution = 0
				} else if m.selectedSolution > 0 {
					m.selectedSolution--
				}
			}

		case "down":
			if m.state == StateShowingSolutions {
				if m.selectedSolution == -1 {
					m.selectedSolution = 0
				} else if m.selectedSolution < len(m.solutions)-1 {
					m.selectedSolution++
				}
			}

		case "backspace":
			if m.state == StateInput && len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			} else if m.state == StateShowingSolutions {
				// User is typing again - go back to input mode
				m.state = StateInput
				m.selectedSolution = -1
			}

		default:
			// Handle regular character input
			if m.state == StateInput {
				m.input += msg.String()
			} else if m.state == StateShowingSolutions {
				// User started typing again - go back to input mode
				m.state = StateInput
				m.selectedSolution = -1
				m.input = msg.String()
			}
		}

	case apiResponseMsg:
		m.solutions = msg.solutions
		m.state = StateShowingSolutions
		m.selectedSolution = -1
		return m, nil

	case apiErrorMsg:
		m.history = append(m.history, errorStyle.Render("Error: "+msg.err.Error()))
		m.state = StateInput
		return m, nil

	case loadingTickMsg:
		if m.state == StateLoading {
			m.loadingFrame++
			return m, tickLoading()
		}
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("idk - AI Command Helper"))
	b.WriteString("\n\n")

	// Show history
	if len(m.history) > 0 {
		for _, h := range m.history {
			if strings.HasPrefix(h, ">") {
				b.WriteString(promptStyle.Render(h))
			} else {
				b.WriteString(historyStyle.Render(h))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Show current state
	switch m.state {
	case StateInput:
		b.WriteString(promptStyle.Render("> "))
		b.WriteString(m.input)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("█"))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Type your query and press Enter. ESC to exit."))

	case StateLoading:
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		frame := spinner[m.loadingFrame%len(spinner)]
		b.WriteString(loadingStyle.Render(frame + " Loading..."))

	case StateShowingSolutions:
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Solutions:"))
		b.WriteString("\n\n")

		for i, sol := range m.solutions {
			prefix := fmt.Sprintf("%d. ", i+1)
			relevanceIndicator := strings.Repeat("●", sol.Relevance)
			content := fmt.Sprintf("%s %s",
				sol.Command,
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(relevanceIndicator))

			if i == m.selectedSolution {
				b.WriteString(selectedSolutionStyle.Render("► " + prefix + content))
			} else {
				b.WriteString(solutionStyle.Render(prefix + content))
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		if m.selectedSolution >= 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to execute, ↑↓ to navigate, or type to continue conversation. ESC to exit."))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Use ↑↓ to select and Enter to execute, or type to continue conversation. ESC to exit."))
		}
	}

	b.WriteString("\n")
	return b.String()
}

// Loading tick message
type loadingTickMsg time.Time

func tickLoading() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return loadingTickMsg(t)
	})
}

// API call to get command suggestions using structured output
func callAPI(prompt string, client *openai.Client, config *ai.Config) tea.Cmd {
	return func() tea.Msg {
		// Check if client is available
		if client == nil || config == nil {
			return apiErrorMsg{err: fmt.Errorf("AI client not configured. Please set OPENAI_API_KEY environment variable")}
		}

		// Create context with timeout to handle network issues
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var result CommandSolutions
		systemPrompt := "You are a helpful command-line assistant. Provide up to 3 specific, working command snippets. Each solution should only include the exact command and a relevance rating (3=most relevant, 1=least relevant). Provide fewer solutions if 1-2 commands are sufficient."

		err := ai.StructuredQueryFromEnv(ctx, prompt, systemPrompt, &result)
		if err != nil {
			// Provide more specific error messages
			if ctx.Err() == context.DeadlineExceeded {
				return apiErrorMsg{err: fmt.Errorf("request timed out after 30 seconds. Please check your network connection")}
			}
			if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "Forbidden") {
				return apiErrorMsg{err: fmt.Errorf("authentication failed. Please check your OPENAI_API_KEY")}
			}
			if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
				return apiErrorMsg{err: fmt.Errorf("rate limit exceeded. Please wait a moment and try again")}
			}
			if strings.Contains(err.Error(), "500") || strings.Contains(err.Error(), "502") || strings.Contains(err.Error(), "503") {
				return apiErrorMsg{err: fmt.Errorf("server error. Please try again later")}
			}
			return apiErrorMsg{err: fmt.Errorf("API error: %v", err)}
		}

		// Validate solutions
		if len(result.Solutions) == 0 {
			return apiErrorMsg{err: fmt.Errorf("no solutions found. Please try rephrasing your query")}
		}

		// Limit to 3 solutions and validate
		validSolutions := make([]CommandSolution, 0, 3)
		for i, sol := range result.Solutions {
			if i >= 3 {
				break
			}
			if strings.TrimSpace(sol.Command) != "" {
				validSolutions = append(validSolutions, sol)
			}
		}

		if len(validSolutions) == 0 {
			return apiErrorMsg{err: fmt.Errorf("no valid solutions found. Please try rephrasing your query")}
		}

		return apiResponseMsg{solutions: validSolutions}
	}
}

// Execute solution - types out the selected command
func executeSolution(solution string) tea.Cmd {
	return func() tea.Msg {
		// Validate solution before executing
		if strings.TrimSpace(solution) == "" {
			return nil
		}

		// Clear screen
		fmt.Print("\033[H\033[2J")

		// Type out the solution character by character
		for _, char := range solution {
			fmt.Print(string(char))
			time.Sleep(30 * time.Millisecond)
		}

		fmt.Println()
		return nil
	}
}

func readPipedInput(maxLen int) string {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		limitedReader := io.LimitReader(os.Stdin, int64(maxLen+1))
		data, err := io.ReadAll(limitedReader)
		if err != nil {
			return ""
		}

		content := strings.TrimSpace(string(data))

		if len(content) > maxLen {
			content = content[:maxLen] + "\n... (truncated)"
		}

		return content
	}

	return ""
}

func main() {
	pipedContext := readPipedInput(400)

	if pipedContext != "" {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			fmt.Printf("Error: Cannot open /dev/tty for interactive input: %v\n", err)
			os.Exit(1)
		}
		os.Stdin = tty
	}

	p := tea.NewProgram(initialModel(pipedContext))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
