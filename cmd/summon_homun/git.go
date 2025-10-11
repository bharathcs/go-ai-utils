package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

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
