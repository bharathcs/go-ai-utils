package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
)

func TestLoadInstructions_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	instructionsPath := filepath.Join(tmpDir, "instructions.md")

	ta := textarea.New()
	m := &model{
		instructionsPath: instructionsPath,
		textarea:         ta,
	}

	err := m.loadInstructions()
	if err != nil {
		t.Fatalf("loadInstructions failed: %v", err)
	}

	content := m.textarea.Value()
	if content != defaultInstructionsTemplate {
		t.Errorf("Expected default template, got: %s", content)
	}
}

func TestLoadInstructions_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	instructionsPath := filepath.Join(tmpDir, "instructions.md")

	testContent := "# Test Instructions\n\nThis is a test."
	err := os.WriteFile(instructionsPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	ta := textarea.New()
	m := &model{
		instructionsPath: instructionsPath,
		textarea:         ta,
	}

	err = m.loadInstructions()
	if err != nil {
		t.Fatalf("loadInstructions failed: %v", err)
	}

	content := m.textarea.Value()
	if content != testContent {
		t.Errorf("Expected %q, got %q", testContent, content)
	}
}

func TestSaveInstructions(t *testing.T) {
	tmpDir := t.TempDir()
	instructionsPath := filepath.Join(tmpDir, "instructions.md")

	ta := textarea.New()
	testContent := "# Saved Instructions\n\nThis should be saved."
	ta.SetValue(testContent)

	m := model{
		instructionsPath: instructionsPath,
		textarea:         ta,
	}

	err := m.saveInstructions()
	if err != nil {
		t.Fatalf("saveInstructions failed: %v", err)
	}

	savedContent, err := os.ReadFile(instructionsPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(savedContent) != testContent {
		t.Errorf("Expected %q, got %q", testContent, string(savedContent))
	}
}
