package service

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ferdikt/taskmd-cli/internal/store"
	"github.com/ferdikt/taskmd-cli/internal/taskfile"
)

type Service struct {
	store *store.Store
	now   func() time.Time
}

type Filters struct {
	Status      taskfile.Status
	HasStatus   bool
	Label       string
	Priority    taskfile.Priority
	HasPriority bool
}

type AddInput struct {
	Title    string   `json:"title"`
	Priority string   `json:"priority,omitempty"`
	Labels   []string `json:"labels,omitempty"`
	Notes    string   `json:"notes,omitempty"`
}

type EditInput struct {
	ID       string    `json:"id"`
	Title    *string   `json:"title,omitempty"`
	Status   *string   `json:"status,omitempty"`
	Priority *string   `json:"priority,omitempty"`
	Labels   *[]string `json:"labels,omitempty"`
	Notes    *string   `json:"notes,omitempty"`
}

func New() *Service {
	return &Service{
		store: store.New(),
		now:   time.Now,
	}
}

func (s *Service) ResolvePath(cwd, override string) (string, error) {
	return store.ResolveTaskFilePath(cwd, override)
}

func (s *Service) ResolveInitPath(cwd, override string) (string, error) {
	return store.ResolveInitPath(cwd, override)
}

func (s *Service) Init(path string, force bool) error {
	return s.store.Init(path, force)
}

func (s *Service) Validate(path string) error {
	return s.store.Validate(path)
}

func (s *Service) Format(path string) error {
	return s.store.Rewrite(path, func(doc *taskfile.Document) error {
		return nil
	})
}

func (s *Service) List(path string, filters Filters) ([]*taskfile.Task, error) {
	doc, err := s.store.Load(path)
	if err != nil {
		return nil, err
	}
	all := doc.AllTasks()
	out := make([]*taskfile.Task, 0, len(all))
	for _, task := range all {
		if filters.HasStatus && task.Status != filters.Status {
			continue
		}
		if filters.HasPriority && task.Priority != filters.Priority {
			continue
		}
		if filters.Label != "" && !slices.Contains(task.Labels, strings.ToLower(filters.Label)) {
			continue
		}
		out = append(out, task)
	}
	return out, nil
}

func (s *Service) Show(path, id string) (*taskfile.Task, error) {
	doc, err := s.store.Load(path)
	if err != nil {
		return nil, err
	}
	task, _, _ := doc.FindTask(id)
	if task == nil {
		return nil, fmt.Errorf("task %s not found", id)
	}
	return task, nil
}

func (s *Service) Add(path string, input AddInput) (*taskfile.Task, error) {
	var created *taskfile.Task
	err := s.store.Rewrite(path, func(doc *taskfile.Document) error {
		priority, err := taskfile.ParsePriority(input.Priority)
		if err != nil {
			return err
		}
		title := strings.TrimSpace(input.Title)
		if title == "" {
			return fmt.Errorf("title cannot be empty")
		}
		now := s.now().Format(time.RFC3339)
		ts, _ := time.Parse(time.RFC3339, now)
		created = &taskfile.Task{
			ID:        doc.NextID(),
			Title:     title,
			Status:    taskfile.StatusTodo,
			Priority:  priority,
			Labels:    taskfile.NormalizeLabels(input.Labels),
			Notes:     strings.TrimSpace(input.Notes),
			CreatedAt: ts,
			UpdatedAt: ts,
		}
		doc.AppendTask(created)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) Edit(path string, patch EditInput) (*taskfile.Task, error) {
	var updated *taskfile.Task
	err := s.store.Rewrite(path, func(doc *taskfile.Document) error {
		task, _, _ := doc.FindTask(patch.ID)
		if task == nil {
			return fmt.Errorf("task %s not found", patch.ID)
		}
		if patch.Title != nil {
			title := strings.TrimSpace(*patch.Title)
			if title == "" {
				return fmt.Errorf("title cannot be empty")
			}
			task.Title = title
		}
		if patch.Priority != nil {
			priority, err := taskfile.ParsePriority(*patch.Priority)
			if err != nil {
				return err
			}
			task.Priority = priority
		}
		if patch.Labels != nil {
			task.Labels = taskfile.NormalizeLabels(*patch.Labels)
		}
		if patch.Notes != nil {
			task.Notes = strings.TrimSpace(*patch.Notes)
		}
		if patch.Status != nil {
			status, err := taskfile.ParseStatus(*patch.Status)
			if err != nil {
				return err
			}
			task, err = doc.MoveTask(task.ID, status)
			if err != nil {
				return err
			}
		}
		task.UpdatedAt = s.now()
		updated = task
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) SetStatus(path string, status taskfile.Status, ids []string) error {
	return s.store.Rewrite(path, func(doc *taskfile.Document) error {
		for _, id := range ids {
			task, err := doc.MoveTask(id, status)
			if err != nil {
				return err
			}
			task.UpdatedAt = s.now()
		}
		return nil
	})
}

func (s *Service) Remove(path string, ids []string) error {
	return s.store.Rewrite(path, func(doc *taskfile.Document) error {
		for _, id := range ids {
			if _, err := doc.RemoveTask(id); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) Next(path string) (*taskfile.Task, error) {
	doc, err := s.store.Load(path)
	if err != nil {
		return nil, err
	}
	if len(doc.Todo) == 0 {
		return nil, nil
	}
	best := doc.Todo[0]
	bestRank := taskfile.PriorityRank(best.Priority)
	for _, task := range doc.Todo[1:] {
		rank := taskfile.PriorityRank(task.Priority)
		if rank < bestRank {
			best = task
			bestRank = rank
		}
	}
	return best, nil
}

func (s *Service) BulkAdd(path string, inputs []AddInput) error {
	return s.store.Rewrite(path, func(doc *taskfile.Document) error {
		now := s.now()
		for _, input := range inputs {
			title := strings.TrimSpace(input.Title)
			if title == "" {
				return fmt.Errorf("title cannot be empty")
			}
			priority, err := taskfile.ParsePriority(input.Priority)
			if err != nil {
				return err
			}
			task := &taskfile.Task{
				ID:        doc.NextID(),
				Title:     title,
				Status:    taskfile.StatusTodo,
				Priority:  priority,
				Labels:    taskfile.NormalizeLabels(input.Labels),
				Notes:     strings.TrimSpace(input.Notes),
				CreatedAt: now,
				UpdatedAt: now,
			}
			doc.AppendTask(task)
		}
		return nil
	})
}

func (s *Service) BulkEdit(path string, patches []EditInput) error {
	return s.store.Rewrite(path, func(doc *taskfile.Document) error {
		for _, patch := range patches {
			task, _, _ := doc.FindTask(patch.ID)
			if task == nil {
				return fmt.Errorf("task %s not found", patch.ID)
			}
			if patch.Title != nil {
				title := strings.TrimSpace(*patch.Title)
				if title == "" {
					return fmt.Errorf("title cannot be empty")
				}
				task.Title = title
			}
			if patch.Priority != nil {
				priority, err := taskfile.ParsePriority(*patch.Priority)
				if err != nil {
					return err
				}
				task.Priority = priority
			}
			if patch.Labels != nil {
				task.Labels = taskfile.NormalizeLabels(*patch.Labels)
			}
			if patch.Notes != nil {
				task.Notes = strings.TrimSpace(*patch.Notes)
			}
			if patch.Status != nil {
				status, err := taskfile.ParseStatus(*patch.Status)
				if err != nil {
					return err
				}
				task, err = doc.MoveTask(task.ID, status)
				if err != nil {
					return err
				}
			}
			task.UpdatedAt = s.now()
		}
		return nil
	})
}

func (s *Service) BulkRemove(path string, ids []string) error {
	return s.Remove(path, ids)
}

func DecodeBulkAdd(data []byte) ([]AddInput, error) {
	var inputs []AddInput
	if err := json.Unmarshal(data, &inputs); err != nil {
		return nil, err
	}
	return inputs, nil
}

func DecodeBulkEdit(data []byte) ([]EditInput, error) {
	var inputs []EditInput
	if err := json.Unmarshal(data, &inputs); err != nil {
		return nil, err
	}
	return inputs, nil
}

func DecodeBulkRemove(data []byte) ([]string, error) {
	var ids []string
	if err := json.Unmarshal(data, &ids); err == nil {
		return ids, nil
	}
	var objects []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &objects); err != nil {
		return nil, err
	}
	ids = make([]string, 0, len(objects))
	for _, obj := range objects {
		ids = append(ids, obj.ID)
	}
	return ids, nil
}

func ReadInput(path string) ([]byte, error) {
	if path == "-" {
		return os.ReadFile("/dev/stdin")
	}
	return os.ReadFile(path)
}
