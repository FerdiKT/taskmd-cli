package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ferdikt/taskmd-cli/internal/taskfile"
)

func PrintTaskJSON(w io.Writer, task *taskfile.Task) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(task)
}

func PrintTasksJSON(w io.Writer, tasks []*taskfile.Task) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tasks)
}

func PrintTaskHuman(w io.Writer, task *taskfile.Task) error {
	if task == nil {
		_, err := fmt.Fprintln(w, "No task found.")
		return err
	}
	_, err := fmt.Fprintf(
		w,
		"%s\n  title: %s\n  status: %s\n  priority: %s\n  labels: %s\n  created: %s\n  updated: %s\n",
		task.ID,
		task.Title,
		task.Status,
		task.Priority,
		strings.Join(task.Labels, ", "),
		task.CreatedAt.Format("2006-01-02 15:04:05 -07:00"),
		task.UpdatedAt.Format("2006-01-02 15:04:05 -07:00"),
	)
	if err != nil {
		return err
	}
	if strings.TrimSpace(task.Notes) == "" {
		return nil
	}
	_, err = fmt.Fprintf(w, "  notes:\n%s\n", indentBlock(task.Notes, "    "))
	return err
}

func PrintTasksHuman(w io.Writer, tasks []*taskfile.Task) error {
	if len(tasks) == 0 {
		_, err := fmt.Fprintln(w, "No tasks found.")
		return err
	}
	lastStatus := taskfile.Status("")
	for _, task := range tasks {
		if task.Status != lastStatus {
			if lastStatus != "" {
				if _, err := fmt.Fprintln(w); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(w, task.Status.Heading()); err != nil {
				return err
			}
			lastStatus = task.Status
		}
		labelSuffix := ""
		if len(task.Labels) > 0 {
			labelSuffix = " {" + strings.Join(task.Labels, ", ") + "}"
		}
		if _, err := fmt.Fprintf(w, "  %s [%s] %s%s\n", task.ID, task.Priority, task.Title, labelSuffix); err != nil {
			return err
		}
	}
	return nil
}

func indentBlock(value, prefix string) string {
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
