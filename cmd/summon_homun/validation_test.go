package main

import "testing"

func TestIsValidBranchName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid alphanumeric", "feature123", true},
		{"valid with hyphen", "feature-branch", true},
		{"valid with underscore", "feature_branch", true},
		{"valid mixed", "feature-123_test", true},
		{"invalid with space", "feature branch", false},
		{"invalid with slash", "feature/branch", false},
		{"invalid with special chars", "feature@branch", false},
		{"invalid with dot", "feature.branch", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidBranchName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidBranchName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
