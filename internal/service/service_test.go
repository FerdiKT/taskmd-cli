package service

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ferdikt/taskmd-cli/internal/taskfile"
)

func TestServiceLifecycleAndNext(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "docs", "Task.md")

	svc := New()
	svc.now = func() time.Time {
		return time.Date(2026, 4, 9, 14, 30, 0, 0, time.FixedZone("+03", 3*60*60))
	}

	if err := svc.Init(path, false); err != nil {
		t.Fatal(err)
	}

	first, err := svc.Add(path, AddInput{Title: "Write parser", Priority: "p2", Labels: []string{"cli"}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Add(path, AddInput{Title: "Ship README", Priority: "p1"})
	if err != nil {
		t.Fatal(err)
	}

	next, err := svc.Next(path)
	if err != nil {
		t.Fatal(err)
	}
	if next.ID != second.ID {
		t.Fatalf("expected %s to be next, got %s", second.ID, next.ID)
	}

	newNotes := "Parser skeleton"
	newLabels := []string{}
	edited, err := svc.Edit(path, EditInput{
		ID:       first.ID,
		Notes:    &newNotes,
		Labels:   &newLabels,
		Priority: stringPtr("p1"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if edited.Priority != taskfile.PriorityP1 || edited.Notes != "Parser skeleton" || len(edited.Labels) != 0 {
		t.Fatalf("unexpected edited task: %#v", edited)
	}

	if err := svc.SetStatus(path, taskfile.StatusInProgress, []string{first.ID}); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetStatus(path, taskfile.StatusDone, []string{first.ID}); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetStatus(path, taskfile.StatusTodo, []string{first.ID}); err != nil {
		t.Fatal(err)
	}

	task, err := svc.Show(path, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != taskfile.StatusTodo {
		t.Fatalf("expected reopened task to be todo, got %s", task.Status)
	}

	if err := svc.Format(path); err != nil {
		t.Fatal(err)
	}
	if err := svc.Validate(path); err != nil {
		t.Fatal(err)
	}
}

func TestBulkAndConcurrentWrites(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "docs", "Task.md")

	svc := New()
	if err := svc.Init(path, false); err != nil {
		t.Fatal(err)
	}

	if err := svc.BulkAdd(path, []AddInput{
		{Title: "A", Priority: "p3"},
		{Title: "B", Priority: "p1"},
	}); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := taskfile.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Todo) != 2 {
		t.Fatalf("expected 2 tasks after bulk add, got %d", len(doc.Todo))
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, title := range []string{"Concurrent 1", "Concurrent 2"} {
		wg.Add(1)
		go func(title string) {
			defer wg.Done()
			_, err := svc.Add(path, AddInput{Title: title, Priority: "p2"})
			errs <- err
		}(title)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	raw, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	doc, err = taskfile.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Todo) != 4 {
		t.Fatalf("expected 4 tasks after concurrent add, got %d", len(doc.Todo))
	}
}

func stringPtr(value string) *string {
	return &value
}
