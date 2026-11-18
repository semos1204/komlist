// Package boardcfg persists the custom titles of the board columns, the way
// sprint dashboards let you rename their columns. Titles live in a small
// JSON file next to the task store (~/.komlist/columns.json), independent
// of the storage backend, so renaming a column never touches a task.
//
// A column is identified by a key: "all", "today" and "overdue" for the
// three fixed columns, "+<tag>" for a tag column (the same +tag notation
// the board cards use).
package boardcfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Keys of the fixed board columns.
const (
	KeyAll     = "all"
	KeyToday   = "today"
	KeyOverdue = "overdue"
)

// FixedKeys lists the fixed columns in board order.
var FixedKeys = []string{KeyAll, KeyToday, KeyOverdue}

// ErrInvalidKey is returned for a key that names no column.
var ErrInvalidKey = errors.New("invalid column")

// ErrEmptyTitle is returned when renaming a column to a blank title.
var ErrEmptyTitle = errors.New("column title must not be empty")

// TagKey returns the column key of a tag.
func TagKey(tag string) string { return "+" + tag }

// ValidateKey accepts "all", "today", "overdue" and "+<tag>".
func ValidateKey(key string) error {
	switch {
	case key == KeyAll, key == KeyToday, key == KeyOverdue:
		return nil
	case strings.HasPrefix(key, "+") && strings.TrimSpace(key[1:]) != "":
		return nil
	default:
		return fmt.Errorf("%w: %q (valid: all, today, overdue, +<tag>)", ErrInvalidKey, key)
	}
}

// Titles maps a column key to its custom title. A missing key means the
// column keeps its default title.
type Titles map[string]string

// Title returns the custom title of key, or def when none is set.
func (t Titles) Title(key, def string) string {
	if v, ok := t[key]; ok && v != "" {
		return v
	}
	return def
}

// Store loads and saves column titles.
type Store interface {
	Load() (Titles, error)
	Save(Titles) error
}

// Rename validates key and title, then stores the new title.
func Rename(s Store, key, title string) error {
	if err := ValidateKey(key); err != nil {
		return err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrEmptyTitle
	}
	t, err := s.Load()
	if err != nil {
		return err
	}
	t[key] = title
	return s.Save(t)
}

// Reset drops the custom title of key, restoring the default one.
func Reset(s Store, key string) error {
	if err := ValidateKey(key); err != nil {
		return err
	}
	t, err := s.Load()
	if err != nil {
		return err
	}
	delete(t, key)
	return s.Save(t)
}

// File is a Store backed by a JSON file.
type File struct{ path string }

// NewFile returns a Store reading and writing path. The file is created on
// the first Save; a missing file loads as no custom titles.
func NewFile(path string) *File { return &File{path: path} }

// Load reads the titles; a missing file yields an empty set.
func (f *File) Load() (Titles, error) {
	data, err := os.ReadFile(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return Titles{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read column titles: %w", err)
	}
	t := Titles{}
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parse %s: %w", f.path, err)
	}
	return t, nil
}

// Save writes the titles atomically (sibling tmp file, then rename).
func (f *File) Save(t Titles) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(f.path), err)
	}
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write column titles: %w", err)
	}
	return os.Rename(tmp, f.path)
}

// Memory is an in-memory Store, for tests.
type Memory struct {
	mu sync.Mutex
	t  Titles
}

// NewMemory returns an empty in-memory Store.
func NewMemory() *Memory { return &Memory{t: Titles{}} }

// Load returns a copy of the stored titles.
func (m *Memory) Load() (Titles, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(Titles, len(m.t))
	for k, v := range m.t {
		out[k] = v
	}
	return out, nil
}

// Save replaces the stored titles with a copy of t.
func (m *Memory) Save(t Titles) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.t = make(Titles, len(t))
	for k, v := range t {
		m.t[k] = v
	}
	return nil
}
