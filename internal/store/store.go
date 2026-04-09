package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ferdikt/taskmd-cli/internal/taskfile"
	"github.com/gofrs/flock"
)

type Store struct{}

func New() *Store {
	return &Store{}
}

func (s *Store) Load(path string) (*taskfile.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return taskfile.Parse(data)
}

func (s *Store) Validate(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return taskfile.Validate(data)
}

func (s *Store) Init(path string, force bool) error {
	lock, err := acquireLock(path)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("%s already exists", path)
	}
	doc := taskfile.NewDocument()
	return atomicWrite(path, taskfile.Render(doc))
}

func (s *Store) Rewrite(path string, mutate func(*taskfile.Document) error) error {
	lock, err := acquireLock(path)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	doc, err := taskfile.Parse(data)
	if err != nil {
		return err
	}
	if err := mutate(doc); err != nil {
		return err
	}
	return atomicWrite(path, taskfile.Render(doc))
}

func acquireLock(path string) (*flock.Flock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	lock := flock.New(path + ".lock")
	locked, err := lock.TryLock()
	if err != nil {
		return nil, err
	}
	if !locked {
		if err := lock.Lock(); err != nil {
			return nil, err
		}
	}
	return lock, nil
}

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".taskmd-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
