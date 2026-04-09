package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTaskFilePathFindsNearestAncestor(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "docs", "Task.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("# Task\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveTaskFilePath(nested, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != target {
		t.Fatalf("expected %s, got %s", target, got)
	}
}

func TestResolveInitPathPrefersGitRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "app", "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveInitPath(nested, "")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, RelativeTaskPath)
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
