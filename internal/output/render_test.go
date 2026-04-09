package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ferdikt/taskmd-cli/internal/taskfile"
)

func TestPrintTasksJSONContract(t *testing.T) {
	task := &taskfile.Task{
		ID:        "T001",
		Title:     "Write docs",
		Status:    taskfile.StatusTodo,
		Priority:  taskfile.PriorityP1,
		Labels:    []string{"docs"},
		Notes:     "Hello",
		CreatedAt: time.Date(2026, 4, 9, 14, 30, 0, 0, time.FixedZone("+03", 3*60*60)),
		UpdatedAt: time.Date(2026, 4, 9, 14, 30, 0, 0, time.FixedZone("+03", 3*60*60)),
	}
	var buf bytes.Buffer
	if err := PrintTasksJSON(&buf, []*taskfile.Task{task}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, fragment := range []string{`"id": "T001"`, `"status": "todo"`, `"priority": "p1"`} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected JSON to contain %s, got %s", fragment, got)
		}
	}
}
