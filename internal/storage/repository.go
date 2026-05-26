// Package storage defines the persistence contract for Tasks and provides
// reference implementations (in-memory and JSON file).
package storage

import (
	"context"
	"errors"

	"github.com/semos1204/komlist/internal/task"
)

// ErrNotFound is returned when a Task cannot be located.
var ErrNotFound = errors.New("task not found")

// Repository persists Tasks. Implementations must be safe for use from a
// single process; multi-process concurrency is implementation-specific.
type Repository interface {
	List(ctx context.Context) ([]task.Task, error)
	Get(ctx context.Context, id int) (task.Task, error)
	Create(ctx context.Context, t task.Task) (task.Task, error)
	Update(ctx context.Context, t task.Task) error
	Delete(ctx context.Context, id int) error
}
