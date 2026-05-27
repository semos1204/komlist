package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (no cgo), registers "sqlite"

	"github.com/semos1204/komlist/internal/task"
)

// SQLiteRepository persists tasks in a SQLite database. It is an alternative
// to JSONRepository, selectable at runtime, demonstrating that the
// storage.Repository port admits multiple backends without touching the
// service or CLI layers.
//
// Scalar fields map to native columns; slice fields (tags, notes, depends_on)
// are stored as JSON text. SQLite handles its own concurrency, so no separate
// interprocess lock is needed.
type SQLiteRepository struct {
	db *sql.DB
}

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS tasks (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	title      TEXT NOT NULL,
	status     TEXT NOT NULL,
	priority   TEXT,
	tags       TEXT,
	notes      TEXT,
	depends_on TEXT,
	due_at     TEXT,
	recur      TEXT,
	created_at TEXT,
	updated_at TEXT
);`

const sqliteColumns = `id, title, status, priority, tags, notes, depends_on, due_at, recur, created_at, updated_at`

// NewSQLite opens (creating if needed) the SQLite database at path. Missing
// parent directories are created with mode 0o755.
func NewSQLite(path string) (*SQLiteRepository, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create parent directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(sqliteSchema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return &SQLiteRepository{db: db}, nil
}

// Close closes the underlying database handle.
func (r *SQLiteRepository) Close() error { return r.db.Close() }

// List implements Repository.
func (r *SQLiteRepository) List(ctx context.Context) ([]task.Task, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+sqliteColumns+" FROM tasks ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []task.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Get implements Repository.
func (r *SQLiteRepository) Get(ctx context.Context, id int) (task.Task, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+sqliteColumns+" FROM tasks WHERE id = ?", id)
	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return task.Task{}, ErrNotFound
	}
	return t, err
}

// Create implements Repository, assigning the auto-increment ID.
func (r *SQLiteRepository) Create(ctx context.Context, t task.Task) (task.Task, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (title, status, priority, tags, notes, depends_on, due_at, recur, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Title, string(t.Status), string(t.Priority),
		marshalJSON(t.Tags), marshalJSON(t.Notes), marshalJSON(t.DependsOn),
		dueValue(t.DueAt), string(t.Recur), timeValue(t.CreatedAt), timeValue(t.UpdatedAt))
	if err != nil {
		return task.Task{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return task.Task{}, err
	}
	t.ID = int(id)
	return t, nil
}

// Update implements Repository.
func (r *SQLiteRepository) Update(ctx context.Context, t task.Task) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET title=?, status=?, priority=?, tags=?, notes=?, depends_on=?, due_at=?, recur=?, created_at=?, updated_at=?
		 WHERE id=?`,
		t.Title, string(t.Status), string(t.Priority),
		marshalJSON(t.Tags), marshalJSON(t.Notes), marshalJSON(t.DependsOn),
		dueValue(t.DueAt), string(t.Recur), timeValue(t.CreatedAt), timeValue(t.UpdatedAt), t.ID)
	if err != nil {
		return err
	}
	return notFoundIfZero(res)
}

// Delete implements Repository.
func (r *SQLiteRepository) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}
	return notFoundIfZero(res)
}

func notFoundIfZero(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface{ Scan(dest ...any) error }

func scanTask(s scanner) (task.Task, error) {
	var (
		t                       task.Task
		status, priority, recur string
		tags, notes, deps       sql.NullString
		due, created, updated   sql.NullString
	)
	if err := s.Scan(&t.ID, &t.Title, &status, &priority, &tags, &notes, &deps, &due, &recur, &created, &updated); err != nil {
		return task.Task{}, err
	}
	t.Status = task.Status(status)
	t.Priority = task.Priority(priority)
	t.Recur = task.Recurrence(recur)
	t.Tags = unmarshalStrings(tags)
	t.Notes = unmarshalStrings(notes)
	t.DependsOn = unmarshalInts(deps)
	t.DueAt = parseDue(due)
	t.CreatedAt = parseTime(created)
	t.UpdatedAt = parseTime(updated)
	return t, nil
}

func marshalJSON[T any](v []T) any {
	if len(v) == 0 {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return string(b)
}

func unmarshalStrings(ns sql.NullString) []string {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	var out []string
	_ = json.Unmarshal([]byte(ns.String), &out)
	return out
}

func unmarshalInts(ns sql.NullString) []int {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	var out []int
	_ = json.Unmarshal([]byte(ns.String), &out)
	return out
}

func dueValue(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func parseDue(ns sql.NullString) *time.Time {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	tt, err := time.Parse(time.RFC3339Nano, ns.String)
	if err != nil {
		return nil
	}
	return &tt
}

func timeValue(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(ns sql.NullString) time.Time {
	if !ns.Valid || ns.String == "" {
		return time.Time{}
	}
	tt, _ := time.Parse(time.RFC3339Nano, ns.String)
	return tt
}
