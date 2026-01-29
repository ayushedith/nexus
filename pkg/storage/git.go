package storage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitRepository struct {
	path string
}

func NewGitRepository(path string) (*GitRepository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("get absolute path: %w", err)
	}

	return &GitRepository{path: absPath}, nil
}

func (r *GitRepository) Init() error {
	cmd := exec.Command("git", "init")
	cmd.Dir = r.path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init: %w", err)
	}

	gitignore := filepath.Join(r.path, ".gitignore")
	content := `.nexus/
.env
*.log
secrets/
`
	if err := os.WriteFile(gitignore, []byte(content), 0644); err != nil {
		return fmt.Errorf("create gitignore: %w", err)
	}

	return nil
}

func (r *GitRepository) IsRepo() bool {
	gitDir := filepath.Join(r.path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func (r *GitRepository) Status() (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status: %w", err)
	}
	return string(output), nil
}

func (r *GitRepository) Add(files ...string) error {
	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	return nil
}

func (r *GitRepository) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = r.path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

func (r *GitRepository) Log(n int) ([]CommitInfo, error) {
	format := "--pretty=format:%H|%an|%ae|%at|%s"
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", n), format)
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 5 {
			continue
		}
		commits = append(commits, CommitInfo{
			Hash:    parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Message: parts[4],
		})
	}

	return commits, nil
}

func (r *GitRepository) Diff(ref1, ref2 string) (string, error) {
	cmd := exec.Command("git", "diff", ref1, ref2)
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return string(output), nil
}

func (r *GitRepository) Branch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *GitRepository) Checkout(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = r.path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout: %w", err)
	}
	return nil
}

func (r *GitRepository) CreateBranch(name string) error {
	cmd := exec.Command("git", "checkout", "-b", name)
	cmd.Dir = r.path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git create branch: %w", err)
	}
	return nil
}

func (r *GitRepository) Pull() error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = r.path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	return nil
}

func (r *GitRepository) Push() error {
	cmd := exec.Command("git", "push")
	cmd.Dir = r.path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}

type CommitInfo struct {
	Hash    string
	Author  string
	Email   string
	Message string
}
