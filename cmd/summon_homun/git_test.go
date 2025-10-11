package main

import (
	"os"
	"testing"
)

func TestGetDefaultRepo(t *testing.T) {
	repo := getDefaultRepo()
	if repo == "" {
		t.Skip("Skipping test: could not determine default repo")
	}

	if len(repo) == 0 {
		t.Error("Expected non-empty repo name")
	}
}

func TestGetDefaultRepoURL(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	url := getDefaultRepoURL()
	if url == "" {
		t.Skip("Skipping test: not in a git repository or no origin remote")
	}
}

func TestGetGitRoot(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	root, err := getGitRoot()
	if err != nil {
		t.Skip("Skipping test: not in a git repository")
	}

	if root == "" {
		t.Error("Expected non-empty git root")
	}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Errorf("Git root directory does not exist: %s", root)
	}
}
