package taskfile

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

const CurrentVersion = 1

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

var orderedStatuses = []Status{StatusTodo, StatusInProgress, StatusDone}

func ValidStatuses() []Status {
	return slices.Clone(orderedStatuses)
}

func (s Status) Heading() string {
	switch s {
	case StatusTodo:
		return "Todo"
	case StatusInProgress:
		return "In Progress"
	case StatusDone:
		return "Done"
	default:
		return string(s)
	}
}

func ParseStatus(value string) (Status, error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "todo":
		return StatusTodo, nil
	case "in_progress":
		return StatusInProgress, nil
	case "done":
		return StatusDone, nil
	default:
		return "", fmt.Errorf("invalid status %q", value)
	}
}

type Priority string

const (
	PriorityP1   Priority = "p1"
	PriorityP2   Priority = "p2"
	PriorityP3   Priority = "p3"
	PriorityNone Priority = "none"
)

func ValidPriorities() []Priority {
	return []Priority{PriorityP1, PriorityP2, PriorityP3, PriorityNone}
}

func ParsePriority(value string) (Priority, error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "p1":
		return PriorityP1, nil
	case "p2":
		return PriorityP2, nil
	case "p3":
		return PriorityP3, nil
	case "", "none":
		return PriorityNone, nil
	default:
		return "", fmt.Errorf("invalid priority %q", value)
	}
}

func PriorityRank(priority Priority) int {
	switch priority {
	case PriorityP1:
		return 0
	case PriorityP2:
		return 1
	case PriorityP3:
		return 2
	default:
		return 3
	}
}

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    Status    `json:"status"`
	Priority  Priority  `json:"priority"`
	Labels    []string  `json:"labels"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Document struct {
	Version    int     `json:"version"`
	Todo       []*Task `json:"todo"`
	InProgress []*Task `json:"in_progress"`
	Done       []*Task `json:"done"`
}

func NewDocument() *Document {
	return &Document{
		Version:    CurrentVersion,
		Todo:       []*Task{},
		InProgress: []*Task{},
		Done:       []*Task{},
	}
}

func (d *Document) AllTasks() []*Task {
	all := make([]*Task, 0, len(d.Todo)+len(d.InProgress)+len(d.Done))
	all = append(all, d.Todo...)
	all = append(all, d.InProgress...)
	all = append(all, d.Done...)
	return all
}

func (d *Document) Section(status Status) *[]*Task {
	switch status {
	case StatusTodo:
		return &d.Todo
	case StatusInProgress:
		return &d.InProgress
	case StatusDone:
		return &d.Done
	default:
		return nil
	}
}

func (d *Document) FindTask(id string) (*Task, Status, int) {
	for _, status := range orderedStatuses {
		section := d.Section(status)
		if section == nil {
			continue
		}
		for idx, task := range *section {
			if task.ID == id {
				return task, status, idx
			}
		}
	}
	return nil, "", -1
}

func (d *Document) RemoveTask(id string) (*Task, error) {
	task, status, idx := d.FindTask(id)
	if task == nil {
		return nil, fmt.Errorf("task %s not found", id)
	}
	section := d.Section(status)
	*section = append((*section)[:idx], (*section)[idx+1:]...)
	return task, nil
}

func (d *Document) MoveTask(id string, target Status) (*Task, error) {
	task, _, _ := d.FindTask(id)
	if task == nil {
		return nil, fmt.Errorf("task %s not found", id)
	}
	if task.Status == target {
		return task, nil
	}
	removed, err := d.RemoveTask(id)
	if err != nil {
		return nil, err
	}
	removed.Status = target
	section := d.Section(target)
	*section = append(*section, removed)
	return removed, nil
}

func (d *Document) AppendTask(task *Task) {
	section := d.Section(task.Status)
	*section = append(*section, task)
}

func (d *Document) NextID() string {
	maxID := 0
	for _, task := range d.AllTasks() {
		if len(task.ID) < 2 || task.ID[0] != 'T' {
			continue
		}
		value, err := strconv.Atoi(task.ID[1:])
		if err == nil && value > maxID {
			maxID = value
		}
	}
	return fmt.Sprintf("T%03d", maxID+1)
}

func NormalizeLabels(labels []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(labels))
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		label = strings.ToLower(label)
		if _, ok := seen[label]; ok {
			continue
		}
		seen[label] = struct{}{}
		out = append(out, label)
	}
	return out
}
