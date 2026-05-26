package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofrs/flock"

	"github.com/semos1204/komlist/internal/task"
)

// DefaultPath returns the default location of the JSON repository file:
// $HOME/.komlist/tasks.json.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	return filepath.Join(home, ".komlist", "tasks.json"), nil
}

// JSONRepository persists Tasks to a single JSON file. Writes are atomic:
// a sibling tmp file is written and renamed in place. The repository keeps
// the entire dataset in memory and rewrites the file on each mutation, which
// is suitable for small CLI workloads (a few thousand tasks).
//
// Interprocess safety is enforced via a sidecar lock file (`<path>.lock`)
// held with flock(2) for the entire lifetime of the repository. Callers
// MUST invoke Close on exit to release the lock; concurrent `kl`
// invocations against the same file therefore serialise rather than
// race-corrupt each other.
type JSONRepository struct {
	path string
	lock *flock.Flock

	mu     sync.Mutex
	tasks  map[int]task.Task
	nextID int
}

type jsonPayload struct {
	Tasks  []task.Task `json:"tasks"`
	NextID int         `json:"next_id"`
}

// NewJSON opens or creates the JSON repository at path. Missing parent
// directories are created with mode 0o755. The call blocks until an
// exclusive interprocess lock is acquired on `<path>.lock`; callers must
// invoke Close to release it.
func NewJSON(path string) (*JSONRepository, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create parent directory: %w", err)
	}
	lock := flock.New(path + ".lock")
	if err := lock.Lock(); err != nil {
		return nil, fmt.Errorf("acquire lock %s: %w", lock.Path(), err)
	}
	r := &JSONRepository{
		path:   path,
		lock:   lock,
		tasks:  make(map[int]task.Task),
		nextID: 1,
	}
	if err := r.load(); err != nil {
		_ = lock.Unlock()
		return nil, err
	}
	return r, nil
}

// Close releases the interprocess lock. It is safe to call multiple times.
func (r *JSONRepository) Close() error {
	if r.lock == nil {
		return nil
	}
	err := r.lock.Unlock()
	r.lock = nil
	return err
}

func (r *JSONRepository) load() error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", r.path, err)
	}
	if len(data) == 0 {
		return nil
	}
	var p jsonPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("parse %s: %w", r.path, err)
	}
	for _, t := range p.Tasks {
		r.tasks[t.ID] = t
	}
	if p.NextID > 0 {
		r.nextID = p.NextID
	}
	return nil
}

func (r *JSONRepository) save() error {
	p := jsonPayload{NextID: r.nextID, Tasks: sortedSnapshot(r.tasks)}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(r.path), filepath.Base(r.path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, r.path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// List implements Repository.
func (r *JSONRepository) List(_ context.Context) ([]task.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return sortedSnapshot(r.tasks), nil
}

// Get implements Repository.
func (r *JSONRepository) Get(_ context.Context, id int) (task.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok {
		return task.Task{}, ErrNotFound
	}
	return t, nil
}

// Create implements Repository.
func (r *JSONRepository) Create(_ context.Context, t task.Task) (task.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t.ID = r.nextID
	r.nextID++
	r.tasks[t.ID] = t
	if err := r.save(); err != nil {
		delete(r.tasks, t.ID)
		r.nextID--
		return task.Task{}, err
	}
	return t, nil
}

// Update implements Repository.
func (r *JSONRepository) Update(_ context.Context, t task.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	prev, ok := r.tasks[t.ID]
	if !ok {
		return ErrNotFound
	}
	r.tasks[t.ID] = t
	if err := r.save(); err != nil {
		r.tasks[t.ID] = prev
		return err
	}
	return nil
}

// Delete implements Repository.
func (r *JSONRepository) Delete(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	prev, ok := r.tasks[id]
	if !ok {
		return ErrNotFound
	}
	delete(r.tasks, id)
	if err := r.save(); err != nil {
		r.tasks[id] = prev
		return err
	}
	return nil
}
