package main

import (
	"fmt"
	"os"
)

const defaultInstructionsTemplate = `# Instructions for Homunculus

## Task Description
[Describe what you want Homunculus to accomplish]

## Context
[Provide any relevant context or background information]

## Requirements
[List specific requirements or constraints]

## Expected Output
[Describe what the final result should look like]
`

func (m *model) loadInstructions() error {
	content, err := os.ReadFile(m.instructionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.textarea.SetValue(defaultInstructionsTemplate)
			return nil
		}
		return fmt.Errorf("failed to read instructions file: %w", err)
	}

	m.textarea.SetValue(string(content))
	return nil
}

func (m model) saveInstructions() error {
	content := m.textarea.Value()
	if err := os.WriteFile(m.instructionsPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write instructions file: %w", err)
	}
	return nil
}
