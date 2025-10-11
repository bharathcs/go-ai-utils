package main

import (
	"os/exec"
	"testing"
)

func TestCheckDockerImage(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not installed, skipping test")
	}

	exists, err := checkDockerImage("nonexistent-image-that-should-not-exist-12345")
	if err != nil {
		t.Fatalf("checkDockerImage failed: %v", err)
	}

	if exists {
		t.Error("Expected nonexistent image to return false")
	}
}
