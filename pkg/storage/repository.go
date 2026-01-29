package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nexusapi/nexus/pkg/collection"
)

type Repository struct {
	basePath string
	git      *GitRepository
	parser   *collection.Parser
}

func NewRepository(basePath string) (*Repository, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("get absolute path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	git, err := NewGitRepository(absPath)
	if err != nil {
		return nil, fmt.Errorf("init git: %w", err)
	}

	if !git.IsRepo() {
		if err := git.Init(); err != nil {
			return nil, fmt.Errorf("git init: %w", err)
		}
	}

	return &Repository{
		basePath: absPath,
		git:      git,
		parser:   collection.NewParser(),
	}, nil
}

func (r *Repository) LoadCollection(name string) (*collection.Collection, error) {
	path := filepath.Join(r.basePath, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join(r.basePath, name+".yaml")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("collection not found: %s", name)
		}
	}

	return r.parser.ParseFile(path)
}

func (r *Repository) SaveCollection(coll *collection.Collection, name string) error {
	path := filepath.Join(r.basePath, name+".yaml")
	return r.parser.SaveFile(coll, path)
}

func (r *Repository) ListCollections() ([]string, error) {
	entries, err := os.ReadDir(r.basePath)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	collections := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext == ".yaml" || ext == ".yml" || ext == ".json" {
			collections = append(collections, entry.Name())
		}
	}

	return collections, nil
}

func (r *Repository) Commit(message string, files ...string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to commit")
	}

	if err := r.git.Add(files...); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	if err := r.git.Commit(message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}

func (r *Repository) History(n int) ([]CommitInfo, error) {
	return r.git.Log(n)
}

func (r *Repository) CurrentBranch() (string, error) {
	return r.git.Branch()
}

func (r *Repository) SwitchBranch(name string) error {
	return r.git.Checkout(name)
}

func (r *Repository) CreateBranch(name string) error {
	return r.git.CreateBranch(name)
}

func (r *Repository) Sync() error {
	if err := r.git.Pull(); err != nil {
		return fmt.Errorf("pull: %w", err)
	}
	if err := r.git.Push(); err != nil {
		return fmt.Errorf("push: %w", err)
	}
	return nil
}
