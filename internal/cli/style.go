package cli

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/semos1204/komlist/internal/task"
)

// Shared lipgloss styles for the board view. ANSI palette indices are used so
// colours adapt to the user's terminal theme. lipgloss/termenv disables colour
// automatically when stdout is not a TTY or NO_COLOR is set.
var (
	groupStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	idStyle     = lipgloss.NewStyle().Faint(true)
	doneStyle   = lipgloss.NewStyle().Faint(true).Strikethrough(true)
	footerStyle = lipgloss.NewStyle().Faint(true)
)

var statusGlyph = map[task.Status]string{
	task.StatusTodo:       "☐",
	task.StatusInProgress: "▶",
	task.StatusBlocked:    "⊘",
	task.StatusDone:       "✔",
}

func statusColor(s task.Status) lipgloss.Color {
	switch s {
	case task.StatusInProgress:
		return lipgloss.Color("11") // yellow
	case task.StatusBlocked:
		return lipgloss.Color("9") // red
	case task.StatusDone:
		return lipgloss.Color("10") // green
	default:
		return lipgloss.Color("7") // grey
	}
}

func priorityColor(p task.Priority) lipgloss.Color {
	switch p {
	case task.PriorityHigh:
		return lipgloss.Color("9")
	case task.PriorityMedium:
		return lipgloss.Color("11")
	default:
		return lipgloss.Color("12")
	}
}

func renderID(id int) string {
	return idStyle.Render(fmt.Sprintf("%d.", id))
}

func renderBullet(s task.Status) string {
	return lipgloss.NewStyle().Foreground(statusColor(s)).Render(statusGlyph[s])
}

func renderPriority(p task.Priority) string {
	return lipgloss.NewStyle().Foreground(priorityColor(p)).Render("·" + string(p))
}

func renderDue(due time.Time) string {
	label := "⚑ " + due.Format(time.DateOnly)
	now := time.Now()
	switch {
	case due.Before(now):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(label)
	case due.Before(now.AddDate(0, 0, 3)):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(label)
	default:
		return idStyle.Render(label)
	}
}

func renderRecur(r task.Recurrence) string {
	return idStyle.Render("⟳" + string(r))
}
