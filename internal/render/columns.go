package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/task"
)

// MinColumnWidth is the narrowest outer width (borders included) a board
// column may take. When the terminal cannot fit every column side by side
// at this width, the board wraps onto additional rows of columns.
const MinColumnWidth = 32

// Column is one titled box of the board view.
type Column struct {
	Title string
	Tasks []task.Task
}

var (
	columnBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: hexFgDimLight, Dark: hexFgDimmer})
	columnTitleStyle = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).
				Foreground(lipgloss.AdaptiveColor{Light: hexFgLight, Dark: hexFg})
	ruleStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: hexFgDimLight, Dark: hexFgDimmer})
)

// Board renders columns as bordered boxes laid out side by side within
// width terminal cells, topydo-style. Each task is a card: a coloured
// urgency bar, its id and status glyph, then title, tags and due label
// wrapped inside the column. Boxes in the same row share the same height.
func Board(cols []Column, blocked map[int]bool, width int, now time.Time) string {
	if len(cols) == 0 {
		return ""
	}
	if width < MinColumnWidth {
		width = MinColumnWidth
	}
	perRow := min(max(width/MinColumnWidth, 1), len(cols))
	inner := width/perRow - 2 // the border takes one cell on each side

	var rows []string
	for start := 0; start < len(cols); start += perRow {
		end := min(start+perRow, len(cols))
		bodies := make([]string, 0, end-start)
		height := 0
		for _, c := range cols[start:end] {
			b := columnBody(c, blocked, inner, now)
			bodies = append(bodies, b)
			height = max(height, lipgloss.Height(b))
		}
		boxes := make([]string, len(bodies))
		for i, b := range bodies {
			boxes[i] = columnBorder.Width(inner).Height(height).Render(b)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, boxes...))
	}
	return strings.Join(rows, "\n")
}

func columnBody(c Column, blocked map[int]bool, inner int, now time.Time) string {
	lines := []string{
		columnTitleStyle.Width(inner).Render(fmt.Sprintf("%s (%d)", c.Title, len(c.Tasks))),
		ruleStyle.Render(strings.Repeat("─", inner)),
	}
	sep := ruleStyle.Render(strings.Repeat("╌", inner))
	for _, t := range c.Tasks {
		lines = append(lines, Card(t, blocked[t.ID], inner, now), sep)
	}
	return strings.Join(lines, "\n")
}

// Card renders one task as a multi-line block exactly width cells wide.
//
//	▌  26 ○ (H) Title wrapped onto
//	▌         several lines
//	▌         +infra +prod
//	▌         due in 2 days
func Card(t task.Task, blocked bool, width int, now time.Time) string {
	done := t.Status == task.StatusDone

	prefix := ID(t.ID) + " " + Bullet(t.Status) + " "
	prefixW := lipgloss.Width(prefix)
	bodyW := max(width-3-prefixW, 8) // bar + gap on the left, one cell of margin on the right
	wrap := lipgloss.NewStyle().Width(bodyW)

	head := []string{}
	if blocked {
		head = append(head, "\U0001F512")
	}
	if t.Priority != "" {
		head = append(head, lipgloss.NewStyle().Bold(true).Foreground(PriorityColor(t.Priority)).
			Render("("+strings.ToUpper(string(t.Priority)[:1])+")"))
	}
	title := t.Title
	if done {
		title = DoneStyle.Render(title)
	}
	head = append(head, title)

	body := []string{wrap.Render(strings.Join(head, " "))}
	if len(t.Tags) > 0 {
		tags := make([]string, len(t.Tags))
		for i, tag := range t.Tags {
			fg := paletteFor(tag).Fg
			if done {
				fg = dimmedTagPalette.Fg
			}
			tags[i] = lipgloss.NewStyle().Foreground(fg).Render("+" + tag)
		}
		body = append(body, wrap.Render(strings.Join(tags, " ")))
	}
	if t.DueAt != nil {
		body = append(body, wrap.Render(dueStyle(now, *t.DueAt, done).Render(DueLabel(now, *t.DueAt))))
	}
	if t.Recur != task.RecurNone {
		body = append(body, Recur(t.Recur))
	}
	block := lipgloss.JoinVertical(lipgloss.Left, body...)

	bar := lipgloss.NewStyle().Foreground(barColor(t, now)).Render("▌")
	barCol := strings.TrimSuffix(strings.Repeat(bar+"\n", lipgloss.Height(block)), "\n")
	return lipgloss.JoinHorizontal(lipgloss.Top, barCol, " ", prefix, block)
}

// DueLabel renders a due date as a sentence relative to now: "due today",
// "due in 4 days", "due 2 days ago".
func DueLabel(now, due time.Time) string {
	switch d := daysBetween(now, due); {
	case d == 0:
		return i18n.T(i18n.KeyDueToday)
	case d == 1:
		return i18n.T(i18n.KeyDueTomorrow)
	case d == -1:
		return i18n.T(i18n.KeyDueYesterday)
	case d > 1:
		return i18n.T(i18n.KeyDueInDays, d)
	default:
		return i18n.T(i18n.KeyDueDaysAgo, -d)
	}
}

// IsOverdue reports whether due falls on a calendar day before now's.
func IsOverdue(now, due time.Time) bool { return daysBetween(now, due) < 0 }

// IsDueToday reports whether due falls on now's calendar day.
func IsDueToday(now, due time.Time) bool { return daysBetween(now, due) == 0 }

func dueStyle(now, due time.Time, done bool) lipgloss.Style {
	s := lipgloss.NewStyle()
	switch d := daysBetween(now, due); {
	case done:
		return FooterStyle
	case d <= 0:
		return s.Foreground(lipgloss.Color(hexPink))
	case d < 3:
		return s.Foreground(lipgloss.Color(hexAmber))
	default:
		return FooterStyle
	}
}

// barColor is the urgency stripe on the left of a card: pink when overdue,
// amber when due today or within three days, green otherwise; completed
// tasks fade out.
func barColor(t task.Task, now time.Time) lipgloss.TerminalColor {
	if t.Status == task.StatusDone {
		return lipgloss.Color(hexFgDimmer)
	}
	if t.DueAt != nil {
		switch d := daysBetween(now, *t.DueAt); {
		case d < 0:
			return lipgloss.Color(hexPink)
		case d < 3:
			return lipgloss.Color(hexAmber)
		}
	}
	return lipgloss.Color(hexGreen)
}
