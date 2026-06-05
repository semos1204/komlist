// Package render holds the shared lipgloss styling used by the board view
// (internal/cli) and the interactive TUI (internal/tui), so both render tasks
// consistently. lipgloss/termenv disables colour automatically when stdout is
// not a TTY or NO_COLOR is set.
//
// The visual language is deliberately minimal:
//
//   - status glyphs progress visually from empty to full: ○ ◐ ⊘ ●
//   - priority is a single bold colour-coded letter (h/m/l), not a word
//   - the due date is a flag plus a short Jan 02 label, coloured red when
//     overdue and yellow when within three days
//   - the cursor in the TUI is a thin vertical accent bar, not a heavy
//     full-row background — so a row's per-element colours stay readable.
package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/semos1204/komlist/internal/task"
)

// UntaggedGroup is the heading used for tasks that carry no tag in
// grouped views.
const UntaggedGroup = "(untagged)"

// Accent is the primary highlight colour, used for the TUI title and cursor
// bar. Adaptive so it stays legible on light and dark terminals.
var Accent = lipgloss.AdaptiveColor{Light: "27", Dark: "39"}

// Shared styles.
var (
	GroupStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "248"})
	IDStyle     = lipgloss.NewStyle().Faint(true)
	DoneStyle   = lipgloss.NewStyle().Faint(true).Strikethrough(true)
	FooterStyle = lipgloss.NewStyle().Faint(true)
)

var statusGlyph = map[task.Status]string{
	task.StatusTodo:       "○",
	task.StatusInProgress: "◐",
	task.StatusBlocked:    "⊘",
	task.StatusDone:       "●",
}

var priorityLetter = map[task.Priority]string{
	task.PriorityHigh:   "h",
	task.PriorityMedium: "m",
	task.PriorityLow:    "l",
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

// ID renders a task ID right-aligned in a 3-column field so single- and
// double-digit ids line up cleanly.
func ID(id int) string { return IDStyle.Render(fmt.Sprintf("%3d", id)) }

// Bullet renders the coloured status glyph.
func Bullet(s task.Status) string {
	return lipgloss.NewStyle().Foreground(StatusColor(s)).Render(statusGlyph[s])
}

// Priority renders the priority as a single bold colour-coded letter
// (h/m/l). Returns the empty string when no priority is set.
func Priority(p task.Priority) string {
	letter, ok := priorityLetter[p]
	if !ok {
		return ""
	}
	return lipgloss.NewStyle().Bold(true).Foreground(PriorityColor(p)).Render(letter)
}

// Due renders a due date as "⚐ Jan 02", coloured red when overdue and
// yellow when within three days; otherwise faint.
func Due(due time.Time) string {
	label := "⚐ " + due.Format("Jan 02")
	now := time.Now()
	switch {
	case due.Before(now):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(label)
	case due.Before(now.AddDate(0, 0, 3)):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(label)
	default:
		return FooterStyle.Render(label)
	}
}

// Recur renders a recurrence marker like "⟳ weekly".
func Recur(r task.Recurrence) string { return FooterStyle.Render("⟳ " + string(r)) }

// TaskLine renders a single task as "bullet id title h ⚐ Jan 02 ⟳ weekly".
// Callers add their own leading indent or cursor marker.
func TaskLine(t task.Task, blocked bool) string {
	parts := []string{Bullet(t.Status), ID(t.ID)}
	if blocked {
		parts = append(parts, "\U0001F512")
	}
	title := t.Title
	if t.Status == task.StatusDone {
		title = DoneStyle.Render(title)
	}
	parts = append(parts, title)
	if prio := Priority(t.Priority); prio != "" {
		parts = append(parts, prio)
	}
	if t.DueAt != nil {
		parts = append(parts, Due(*t.DueAt))
	}
	if t.Recur != task.RecurNone {
		parts = append(parts, Recur(t.Recur))
	}
	return strings.Join(parts, " ")
}
