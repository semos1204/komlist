// Package render holds the shared lipgloss styling used by the board view
// (internal/cli) and the interactive TUI (internal/tui), so both render tasks
// consistently. lipgloss/termenv disables colour automatically when stdout is
// not a TTY or NO_COLOR is set.
package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/semos1204/komlist/internal/task"
)

// Shared styles.
var (
	GroupStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	IDStyle     = lipgloss.NewStyle().Faint(true)
	DoneStyle   = lipgloss.NewStyle().Faint(true).Strikethrough(true)
	FooterStyle = lipgloss.NewStyle().Faint(true)
)

var statusGlyph = map[task.Status]string{
	task.StatusTodo:       "☐",
	task.StatusInProgress: "▶",
	task.StatusBlocked:    "⊘",
	task.StatusDone:       "✔",
}

// StatusColor returns the ANSI palette colour for a status.
func StatusColor(s task.Status) lipgloss.Color {
	switch s {
	case task.StatusInProgress:
		return lipgloss.Color("11")
	case task.StatusBlocked:
		return lipgloss.Color("9")
	case task.StatusDone:
		return lipgloss.Color("10")
	default:
		return lipgloss.Color("7")
	}
}

// PriorityColor returns the ANSI palette colour for a priority.
func PriorityColor(p task.Priority) lipgloss.Color {
	switch p {
	case task.PriorityHigh:
		return lipgloss.Color("9")
	case task.PriorityMedium:
		return lipgloss.Color("11")
	default:
		return lipgloss.Color("12")
	}
}

// ID renders a task ID like "12.".
func ID(id int) string { return IDStyle.Render(fmt.Sprintf("%d.", id)) }

// Bullet renders the coloured status glyph.
func Bullet(s task.Status) string {
	return lipgloss.NewStyle().Foreground(StatusColor(s)).Render(statusGlyph[s])
}

// Priority renders a coloured priority label like "·high".
func Priority(p task.Priority) string {
	return lipgloss.NewStyle().Foreground(PriorityColor(p)).Render("·" + string(p))
}

// Due renders a due date, coloured red when overdue and yellow when within
// three days.
func Due(due time.Time) string {
	label := "⚑ " + due.Format(time.DateOnly)
	now := time.Now()
	switch {
	case due.Before(now):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(label)
	case due.Before(now.AddDate(0, 0, 3)):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(label)
	default:
		return IDStyle.Render(label)
	}
}

// Recur renders a recurrence marker like "⟳weekly".
func Recur(r task.Recurrence) string { return IDStyle.Render("⟳" + string(r)) }

// TaskLine renders a single task as "id. bullet [🔒] title ·prio ⚑due ⟳recur".
// Callers add their own leading indent or cursor.
func TaskLine(t task.Task, blocked bool) string {
	parts := []string{ID(t.ID), Bullet(t.Status)}
	if blocked {
		parts = append(parts, "\U0001F512")
	}
	title := t.Title
	if t.Status == task.StatusDone {
		title = DoneStyle.Render(title)
	}
	parts = append(parts, title)
	if t.Priority != "" {
		parts = append(parts, Priority(t.Priority))
	}
	if t.DueAt != nil {
		parts = append(parts, Due(*t.DueAt))
	}
	if t.Recur != task.RecurNone {
		parts = append(parts, Recur(t.Recur))
	}
	return strings.Join(parts, " ")
}

// TaskLinePlain renders the same content as TaskLine but without any embedded
// styling, so callers can apply a single uniform style (e.g. a selection
// background) over the whole line.
func TaskLinePlain(t task.Task, blocked bool) string {
	parts := []string{fmt.Sprintf("%d.", t.ID), statusGlyph[t.Status]}
	if blocked {
		parts = append(parts, "\U0001F512")
	}
	parts = append(parts, t.Title)
	if t.Priority != "" {
		parts = append(parts, "·"+string(t.Priority))
	}
	if t.DueAt != nil {
		parts = append(parts, "⚑ "+t.DueAt.Format(time.DateOnly))
	}
	if t.Recur != task.RecurNone {
		parts = append(parts, "⟳"+string(t.Recur))
	}
	return strings.Join(parts, " ")
}
