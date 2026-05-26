package storage

import (
	"context"
	"sort"
	"sync"

	"github.com/semos1204/komlist/internal/task"
)

// MemoryRepository is an in-memory Repository. It is the reference
// implementation used for contract testing and can also back a future
// "dry run" mode.
type MemoryRepository struct {
	mu     sync.Mutex
	tasks  map[int]task.Task
	nextID int
}

// NewMemory returns an empty MemoryRepository.
func NewMemory() *MemoryRepository {
	return &MemoryRepository{tasks: make(map[int]task.Task), nextID: 1}
}

// List returns all stored tasks sorted by ID ascending.
func (r *MemoryRepository) List(_ context.Context) ([]task.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return sortedSnapshot(r.tasks), nil
}

// Get returns the Task with the given id or ErrNotFound.
func (r *MemoryRepository) Get(_ context.Context, id int) (task.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok {
		return task.Task{}, ErrNotFound
	}
	return t, nil
}

// Create assigns the next available ID and stores the task.
func (r *MemoryRepository) Create(_ context.Context, t task.Task) (task.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t.ID = r.nextID
	r.nextID++
	r.tasks[t.ID] = t
	return t, nil
}

// Update overwrites the stored Task with the one provided. Returns
// ErrNotFound if no Task with the given ID exists.
func (r *MemoryRepository) Update(_ context.Context, t task.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[t.ID]; !ok {
		return ErrNotFound
	}
	r.tasks[t.ID] = t
	return nil
}

// Delete removes the Task with the given id or returns ErrNotFound.
func (r *MemoryRepository) Delete(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(r.tasks, id)
	return nil
}

func sortedSnapshot(m map[int]task.Task) []task.Task {
	out := make([]task.Task, 0, len(m))
	for _, t := range m {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
