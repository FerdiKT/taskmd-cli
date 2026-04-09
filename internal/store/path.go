package store

import (
	"fmt"
	"os"
	"path/filepath"
)

const RelativeTaskPath = "docs/Task.md"

func ResolveTaskFilePath(cwd, override string) (string, error) {
	if override != "" {
		path, err := filepath.Abs(override)
		if err != nil {
			return "", err
		}
		return path, nil
	}

	dir := cwd
	for {
		candidate := filepath.Join(dir, RelativeTaskPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find %s from %s", RelativeTaskPath, cwd)
}

func ResolveInitPath(cwd, override string) (string, error) {
	if override != "" {
		return ResolveTaskFilePath(cwd, override)
	}
	root, ok := findGitRoot(cwd)
	if ok {
		return filepath.Join(root, RelativeTaskPath), nil
	}
	return filepath.Join(cwd, RelativeTaskPath), nil
}

func findGitRoot(dir string) (string, bool) {
	current := dir
	for {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}
