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
	task.StatusInProgress: "●",
	task.StatusBlocked:    "⊘",
	task.StatusDone:       "✓",
}

// StatusColor returns the ANSI palette colour for a status. Uses the
// 256-colour palette to match a modern, slightly muted look.
func StatusColor(s task.Status) lipgloss.Color {
	switch s {
	case task.StatusInProgress:
		return lipgloss.Color("214") // amber
	case task.StatusBlocked:
		return lipgloss.Color("198") // pink/red
	case task.StatusDone:
		return lipgloss.Color("78") // green
	default:
		return lipgloss.Color("244") // muted grey
	}
}

// PriorityColor returns the ANSI palette colour for a priority. The same
// modern palette as statuses, biased toward warm = important, cool = relaxed.
func PriorityColor(p task.Priority) lipgloss.Color {
	switch p {
	case task.PriorityHigh:
		return lipgloss.Color("198") // hot pink
	case task.PriorityMedium:
		return lipgloss.Color("214") // amber
	default:
		return lipgloss.Color("75") // cool blue
	}
}

// ID renders a task ID right-aligned in a 3-column field so single- and
// double-digit ids line up cleanly.
func ID(id int) string { return IDStyle.Render(fmt.Sprintf("%3d", id)) }

// Bullet renders the coloured status glyph.
func Bullet(s task.Status) string {
	return lipgloss.NewStyle().Foreground(StatusColor(s)).Render(statusGlyph[s])
}

// Priority renders the priority as a single colour-coded filled dot,
// the visual shorthand most modern task UIs use. Returns the empty
// string when no priority is set.
func Priority(p task.Priority) string {
	if p == "" {
		return ""
	}
	return lipgloss.NewStyle().Foreground(PriorityColor(p)).Render("●")
}

// Due renders a due date as a short relative label — "today", "2d",
// "3w" — falling back to "MMM DD" once we're more than a month out (or
// in). Past dates are negative. Colour: pink when overdue, amber when
// within three days, faint otherwise.
func Due(due time.Time) string {
	now := time.Now()
	label := relativeDue(now, due)
	switch {
	case due.Before(now) && !sameDay(due, now):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("198")).Render(label)
	case due.Before(now.AddDate(0, 0, 3)):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(label)
	default:
		return FooterStyle.Render(label)
	}
}

func relativeDue(now, due time.Time) string {
	if sameDay(now, due) {
		return "today"
	}
	days := daysBetween(now, due)
	switch {
	case days >= 1 && days < 7:
		return fmt.Sprintf("%dd", days)
	case days >= 7 && days < 30:
		return fmt.Sprintf("%dw", days/7)
	case days <= -1 && days > -7:
		return fmt.Sprintf("%dd", days)
	case days <= -7 && days > -30:
		return fmt.Sprintf("%dw", days/7)
	default:
		return due.Format("Jan 02")
	}
}

// daysBetween counts whole calendar days between two instants, positive
// when `to` is in the future.
func daysBetween(from, to time.Time) int {
	fromDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())
	return int(toDay.Sub(fromDay).Hours() / 24)
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
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

// TaskLinePlain renders the same content as TaskLine but with no embedded
// styling, so callers can wrap the whole line in a single style (typically
// the selection background bar in the TUI). The leading bullet and ID still
// keep their respective widths, just without colour.
func TaskLinePlain(t task.Task, blocked bool) string {
	parts := []string{statusGlyph[t.Status], fmt.Sprintf("%3d", t.ID)}
	if blocked {
		parts = append(parts, "\U0001F512")
	}
	parts = append(parts, t.Title)
	if t.Priority != "" {
		parts = append(parts, "●") // plain dot — caller may colour it via wrapping style
	}
	if t.DueAt != nil {
		parts = append(parts, relativeDue(time.Now(), *t.DueAt))
	}
	if t.Recur != task.RecurNone {
		parts = append(parts, "⟳ "+string(t.Recur))
	}
	return strings.Join(parts, " ")
}
