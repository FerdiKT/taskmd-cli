package taskfile

import (
	"strings"
	"testing"
	"time"
)

func TestParseAndRenderRoundTrip(t *testing.T) {
	timestamp := time.Date(2026, 4, 9, 14, 30, 0, 0, time.FixedZone("+03", 3*60*60))
	doc := NewDocument()
	doc.Todo = append(doc.Todo, &Task{
		ID:        "T001",
		Title:     "Initialize parser",
		Status:    StatusTodo,
		Priority:  PriorityP1,
		Assignee:  "main-agent",
		Labels:    []string{"cli", "v1"},
		Notes:     "Create the Markdown parser and renderer.\n\n- preserve headings\n- preserve notes",
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	})

	rendered := Render(doc)
	parsed, err := Parse(rendered)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(parsed.Todo) != 1 {
		t.Fatalf("expected 1 todo task, got %d", len(parsed.Todo))
	}
	got := parsed.Todo[0]
	if got.ID != "T001" || got.Title != "Initialize parser" || got.Priority != PriorityP1 || got.Assignee != "main-agent" {
		t.Fatalf("unexpected parsed task: %#v", got)
	}
	if got.Notes != doc.Todo[0].Notes {
		t.Fatalf("notes mismatch:\nwant:\n%s\n\ngot:\n%s", doc.Todo[0].Notes, got.Notes)
	}
}

func TestParseRejectsMissingMetadata(t *testing.T) {
	_, err := Parse([]byte(`# Task

<!-- taskmd:version 1 -->

## Todo

### T001 - Missing metadata

## In Progress

## Done
`))
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "metadata") {
		t.Fatalf("unexpected error: %v", err)
	}
}
