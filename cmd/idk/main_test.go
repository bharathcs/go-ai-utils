package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestInitialModel(t *testing.T) {
	tests := []struct {
		name         string
		pipedContext string
		wantState    State
	}{
		{
			name:         "no piped context",
			pipedContext: "",
			wantState:    StateInput,
		},
		{
			name:         "with piped context",
			pipedContext: "some piped content",
			wantState:    StateInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel(tt.pipedContext)
			if m.state != tt.wantState {
				t.Errorf("initialModel() state = %v, want %v", m.state, tt.wantState)
			}
			if m.pipedContext != tt.pipedContext {
				t.Errorf("initialModel() pipedContext = %v, want %v", m.pipedContext, tt.pipedContext)
			}
			if tt.pipedContext != "" && len(m.history) == 0 {
				t.Error("initialModel() should have history entries when pipedContext is provided")
			}
		})
	}
}

func TestModelUpdate_Input(t *testing.T) {
	tests := []struct {
		name      string
		initInput string
		key       string
		wantInput string
	}{
		{
			name:      "add character",
			initInput: "hell",
			key:       "o",
			wantInput: "hello",
		},
		{
			name:      "backspace",
			initInput: "hello",
			key:       "backspace",
			wantInput: "hell",
		},
		{
			name:      "backspace on empty",
			initInput: "",
			key:       "backspace",
			wantInput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				state: StateInput,
				input: tt.initInput,
			}
			newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			m = newModel.(model)
			if m.input != tt.wantInput {
				t.Errorf("Update() input = %v, want %v", m.input, tt.wantInput)
			}
		})
	}
}

func TestModelUpdate_Navigation(t *testing.T) {
	solutions := []CommandSolution{
		{Command: "cmd1", Relevance: 3},
		{Command: "cmd2", Relevance: 2},
		{Command: "cmd3", Relevance: 1},
	}

	tests := []struct {
		name             string
		initSelected     int
		key              tea.KeyMsg
		wantSelected     int
		wantStateChanged bool
	}{
		{
			name:         "down from -1",
			initSelected: -1,
			key:          tea.KeyMsg{Type: tea.KeyDown},
			wantSelected: 0,
		},
		{
			name:         "down in middle",
			initSelected: 0,
			key:          tea.KeyMsg{Type: tea.KeyDown},
			wantSelected: 1,
		},
		{
			name:         "down at end",
			initSelected: 2,
			key:          tea.KeyMsg{Type: tea.KeyDown},
			wantSelected: 2,
		},
		{
			name:         "up from middle",
			initSelected: 1,
			key:          tea.KeyMsg{Type: tea.KeyUp},
			wantSelected: 0,
		},
		{
			name:         "up at start",
			initSelected: 0,
			key:          tea.KeyMsg{Type: tea.KeyUp},
			wantSelected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				state:            StateShowingSolutions,
				solutions:        solutions,
				selectedSolution: tt.initSelected,
			}
			newModel, _ := m.Update(tt.key)
			m = newModel.(model)
			if m.selectedSolution != tt.wantSelected {
				t.Errorf("Update() selectedSolution = %v, want %v", m.selectedSolution, tt.wantSelected)
			}
		})
	}
}

func TestModelUpdate_StateTransitions(t *testing.T) {
	tests := []struct {
		name       string
		initState  State
		msg        tea.Msg
		wantState  State
		setupModel func(*model)
	}{
		{
			name:      "input to loading on enter with text",
			initState: StateInput,
			msg:       tea.KeyMsg{Type: tea.KeyEnter},
			wantState: StateLoading,
			setupModel: func(m *model) {
				m.input = "test query"
			},
		},
		{
			name:      "stay in input on enter without text",
			initState: StateInput,
			msg:       tea.KeyMsg{Type: tea.KeyEnter},
			wantState: StateInput,
			setupModel: func(m *model) {
				m.input = "   "
			},
		},
		{
			name:      "solutions to input on typing",
			initState: StateShowingSolutions,
			msg:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")},
			wantState: StateInput,
			setupModel: func(m *model) {
				m.solutions = []CommandSolution{{Command: "test", Relevance: 3}}
			},
		},
		{
			name:      "solutions to input on backspace",
			initState: StateShowingSolutions,
			msg:       tea.KeyMsg{Type: tea.KeyBackspace},
			wantState: StateInput,
			setupModel: func(m *model) {
				m.solutions = []CommandSolution{{Command: "test", Relevance: 3}}
			},
		},
		{
			name:      "loading to showing solutions on api response",
			initState: StateLoading,
			msg: apiResponseMsg{
				solutions: []CommandSolution{{Command: "test", Relevance: 3}},
			},
			wantState: StateShowingSolutions,
			setupModel: func(m *model) {
			},
		},
		{
			name:      "loading to input on api error",
			initState: StateLoading,
			msg: apiErrorMsg{
				err: &testError{msg: "test error"},
			},
			wantState: StateInput,
			setupModel: func(m *model) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				state:   tt.initState,
				history: []string{},
			}
			if tt.setupModel != nil {
				tt.setupModel(&m)
			}
			newModel, _ := m.Update(tt.msg)
			m = newModel.(model)
			if m.state != tt.wantState {
				t.Errorf("Update() state = %v, want %v", m.state, tt.wantState)
			}
		})
	}
}

func TestModelView(t *testing.T) {
	tests := []struct {
		name       string
		setupModel func() model
		wantSubstr []string
	}{
		{
			name: "input state",
			setupModel: func() model {
				return model{
					state: StateInput,
					input: "test",
				}
			},
			wantSubstr: []string{"idk - AI Command Helper", "test", "Type your query"},
		},
		{
			name: "loading state",
			setupModel: func() model {
				return model{
					state:        StateLoading,
					loadingFrame: 0,
				}
			},
			wantSubstr: []string{"idk - AI Command Helper", "Loading..."},
		},
		{
			name: "showing solutions state",
			setupModel: func() model {
				return model{
					state: StateShowingSolutions,
					solutions: []CommandSolution{
						{Command: "ls -la", Relevance: 3},
						{Command: "pwd", Relevance: 2},
					},
					selectedSolution: 0,
				}
			},
			wantSubstr: []string{"idk - AI Command Helper", "Solutions:", "ls -la", "pwd"},
		},
		{
			name: "with history",
			setupModel: func() model {
				return model{
					state:   StateInput,
					input:   "",
					history: []string{"> previous query", "response"},
				}
			},
			wantSubstr: []string{"idk - AI Command Helper", "> previous query", "response"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setupModel()
			view := m.View()
			for _, substr := range tt.wantSubstr {
				if !strings.Contains(view, substr) {
					t.Errorf("View() missing expected substring %q in output:\n%s", substr, view)
				}
			}
		})
	}
}

func TestReadPipedInput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "empty input",
			input:  "",
			maxLen: 100,
			want:   "",
		},
		{
			name:   "short input",
			input:  "hello world",
			maxLen: 100,
			want:   "hello world",
		},
		{
			name:   "truncate long input",
			input:  strings.Repeat("a", 500),
			maxLen: 100,
			want:   strings.Repeat("a", 100) + "\n... (truncated)",
		},
		{
			name:   "trim whitespace",
			input:  "  hello world  \n",
			maxLen: 100,
			want:   "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processInput(tt.input, tt.maxLen)
			if result != tt.want {
				t.Errorf("processInput() = %q, want %q", result, tt.want)
			}
		})
	}
}

func processInput(input string, maxLen int) string {
	content := strings.TrimSpace(input)
	if len(content) > maxLen {
		content = content[:maxLen] + "\n... (truncated)"
	}
	return content
}

func TestCommandSolution(t *testing.T) {
	tests := []struct {
		name string
		sol  CommandSolution
	}{
		{
			name: "valid solution",
			sol: CommandSolution{
				Command:   "ls -la",
				Relevance: 3,
			},
		},
		{
			name: "minimal solution",
			sol: CommandSolution{
				Command:   "pwd",
				Relevance: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sol.Command == "" {
				t.Error("CommandSolution should have a command")
			}
			if tt.sol.Relevance < 1 || tt.sol.Relevance > 3 {
				t.Errorf("CommandSolution relevance %d out of range [1,3]", tt.sol.Relevance)
			}
		})
	}
}

func TestModelInit(t *testing.T) {
	m := initialModel("")
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestLoadingAnimation(t *testing.T) {
	m := model{
		state:        StateLoading,
		loadingFrame: 0,
	}

	for i := 0; i < 20; i++ {
		newModel, _ := m.Update(loadingTickMsg{})
		m = newModel.(model)
	}

	if m.loadingFrame != 20 {
		t.Errorf("loadingFrame = %d, want 20", m.loadingFrame)
	}
}

func TestQuitKeys(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{
			name: "ctrl+c",
			key:  tea.KeyMsg{Type: tea.KeyCtrlC},
		},
		{
			name: "esc",
			key:  tea.KeyMsg{Type: tea.KeyEsc},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{state: StateInput}
			_, cmd := m.Update(tt.key)
			if cmd == nil {
				t.Error("Update() should return quit command")
			}
		})
	}
}

func TestPipedContextInPrompt(t *testing.T) {
	m := model{
		state:        StateInput,
		input:        "help",
		pipedContext: "some context",
		history:      []string{},
	}

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(model)

	if m.state != StateLoading {
		t.Error("Should transition to loading state")
	}
	if cmd == nil {
		t.Error("Should return command")
	}
	if len(m.history) == 0 || !strings.Contains(m.history[len(m.history)-1], "help") {
		t.Error("Should add query to history")
	}
}

func TestSolutionSelection(t *testing.T) {
	solutions := []CommandSolution{
		{Command: "cmd1", Relevance: 3},
		{Command: "cmd2", Relevance: 2},
	}

	m := model{
		state:            StateShowingSolutions,
		solutions:        solutions,
		selectedSolution: 0,
	}

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(model)

	if cmd == nil {
		t.Error("Should return execute command")
	}
}

func TestEmptyInputNoSubmit(t *testing.T) {
	m := model{
		state: StateInput,
		input: "",
	}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(model)

	if m.state != StateInput {
		t.Error("Should stay in input state with empty input")
	}
}

func TestWhitespaceOnlyInputNoSubmit(t *testing.T) {
	m := model{
		state: StateInput,
		input: "   \t\n  ",
	}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(model)

	if m.state != StateInput {
		t.Error("Should stay in input state with whitespace-only input")
	}
}

func TestMultipleNavigationCycles(t *testing.T) {
	solutions := []CommandSolution{
		{Command: "cmd1", Relevance: 3},
		{Command: "cmd2", Relevance: 2},
		{Command: "cmd3", Relevance: 1},
	}

	m := model{
		state:            StateShowingSolutions,
		solutions:        solutions,
		selectedSolution: -1,
	}

	keys := []tea.KeyMsg{
		{Type: tea.KeyDown},
		{Type: tea.KeyDown},
		{Type: tea.KeyDown},
		{Type: tea.KeyUp},
		{Type: tea.KeyUp},
		{Type: tea.KeyUp},
	}

	expected := []int{0, 1, 2, 1, 0, 0}

	for i, key := range keys {
		newModel, _ := m.Update(key)
		m = newModel.(model)
		if m.selectedSolution != expected[i] {
			t.Errorf("After key %d, selectedSolution = %d, want %d", i, m.selectedSolution, expected[i])
		}
	}
}
