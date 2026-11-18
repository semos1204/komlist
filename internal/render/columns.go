package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/task"
)

// MinColumnWidth is the narrowest outer width (borders included) a board
// column may take. Callers use it to decide how many columns fit side by
// side; Board itself wraps onto extra rows when given more than that.
const MinColumnWidth = 26

// maxTitleLines caps how many lines a card title may wrap onto; longer
// titles end with an ellipsis.
const maxTitleLines = 2

// Column is one titled box of the board view.
type Column struct {
	Title string
	Tasks []task.Task
}

var (
	columnBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: hexFgDimLight, Dark: hexFgDimmer})
	columnTitleStyle = lipgloss.NewStyle().Align(lipgloss.Center).
				Foreground(lipgloss.AdaptiveColor{Light: hexFgLight, Dark: hexFg})
	ruleStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: hexFgDimLight, Dark: hexFgDimmer})
)

// Board renders columns as bordered boxes laid out side by side within
// width terminal cells, topydo-style. Each task is a card — coloured
// urgency bar, id, status glyph, priority, title, +tags, due label — and
// cards are separated by a dashed rule.
//
// height is the number of terminal lines the whole board may use (0 means
// unlimited, e.g. when piped). When a column holds more cards than fit, it
// ends with "… +N more" instead of growing past the screen.
func Board(cols []Column, blocked map[int]bool, width, height int, now time.Time) string {
	if len(cols) == 0 {
		return ""
	}
	width = max(width, MinColumnWidth)
	perRow := min(max(width/MinColumnWidth, 1), len(cols))
	inner := width/perRow - 2 // the border takes one cell on each side
	nRows := (len(cols) + perRow - 1) / perRow

	// Lines available for cards in each box: the row's share of the
	// height, minus the border (2) and the title + rule (2).
	limit := 0
	if height > 0 {
		limit = max(height/nRows-4, 1)
	}

	var rows []string
	for start := 0; start < len(cols); start += perRow {
		end := min(start+perRow, len(cols))
		bodies := make([]string, 0, end-start)
		boxH := 0
		for _, c := range cols[start:end] {
			b := columnBody(c, blocked, inner, limit, now)
			bodies = append(bodies, b)
			boxH = max(boxH, lipgloss.Height(b))
		}
		boxes := make([]string, len(bodies))
		for i, b := range bodies {
			boxes[i] = columnBorder.Width(inner).Height(boxH).Render(b)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, boxes...))
	}
	return strings.Join(rows, "\n")
}

// columnBody renders the title, the rule and as many cards as fit in
// limit lines (0 = no limit), closing with "… +N more" when some are cut.
func columnBody(c Column, blocked map[int]bool, inner, limit int, now time.Time) string {
	lines := []string{
		columnTitleStyle.Width(inner).Render(fmt.Sprintf("%s (%d)", c.Title, len(c.Tasks))),
		ruleStyle.Render(strings.Repeat("─", inner)),
	}
	sep := ruleStyle.Render(strings.Repeat("-", inner))
	used := 0
	for i, t := range c.Tasks {
		card := Card(t, blocked[t.ID], inner, now)
		need := lipgloss.Height(card) + 1 // card + its dashed rule
		reserve := 0
		if i < len(c.Tasks)-1 {
			reserve = 1 // keep room for the "… +N more" line
		}
		if limit > 0 && used+need+reserve > limit {
			more := FooterStyle.Width(inner).Align(lipgloss.Right).
				Render(i18n.T(i18n.KeyBoardMore, len(c.Tasks)-i))
			lines = append(lines, more)
			break
		}
		lines = append(lines, card, sep)
		used += need
	}
	return strings.Join(lines, "\n")
}

// Card renders one task as a block exactly width cells wide, topydo-style:
//
//	▌  25 ○ (H) Fixer alerte CloudWatch
//	▌         RDS EU Writer CPU (action…
//	▌         +qq +infra
//	▌         due 166 days ago
//
// The title wraps onto two lines at most, tags fit on one line; both end
// with "…" when cut.
func Card(t task.Task, blocked bool, width int, now time.Time) string {
	prefix := ID(t.ID) + " " + Bullet(t.Status) + " "
	bodyW := max(width-3-lipgloss.Width(prefix), 8) // bar + gap on the left, one cell of margin on the right

	body := titleLines(t, blocked, bodyW)
	if tags := tagLine(t, bodyW); tags != "" {
		body = append(body, tags)
	}
	if t.DueAt != nil {
		done := t.Status == task.StatusDone
		body = append(body, dueStyle(now, *t.DueAt, done).Render(ansi.Truncate(DueLabel(now, *t.DueAt), bodyW, "…")))
	}
	if t.Recur != task.RecurNone {
		body = append(body, Recur(t.Recur))
	}
	block := strings.Join(body, "\n")

	bar := lipgloss.NewStyle().Foreground(barColor(t, now)).Render("▌")
	barCol := strings.TrimSuffix(strings.Repeat(bar+"\n", len(body)), "\n")
	return lipgloss.JoinHorizontal(lipgloss.Top, barCol, " ", prefix, block)
}

// titleLines wraps the plain title (with its lock and priority marker) to
// width, keeps maxTitleLines lines, then applies the styling — wrapping
// the plain text first keeps the ellipsis maths free of escape codes.
func titleLines(t task.Task, blocked bool, width int) []string {
	var marker string
	if t.Priority != "" {
		marker = "(" + strings.ToUpper(string(t.Priority)[:1]) + ")"
	}
	head := t.Title
	if marker != "" {
		head = marker + " " + head
	}
	if blocked {
		head = "\U0001F512 " + head
	}
	lines := strings.Split(ansi.Wrap(head, width, ""), "\n")
	if len(lines) > maxTitleLines {
		lines = lines[:maxTitleLines]
		last := strings.TrimRight(lines[maxTitleLines-1], " ")
		lines[maxTitleLines-1] = ansi.Truncate(last, width-1, "") + "…"
	}
	for i, l := range lines {
		l = strings.TrimRight(l, " ")
		if t.Status == task.StatusDone {
			l = DoneStyle.Render(l)
		} else if i == 0 && marker != "" {
			styled := lipgloss.NewStyle().Bold(true).Foreground(PriorityColor(t.Priority)).Render(marker)
			l = strings.Replace(l, marker, styled, 1)
		}
		lines[i] = l
	}
	return lines
}

// tagLine renders +tags on one line of at most width cells, cut with "…".
func tagLine(t task.Task, width int) string {
	if len(t.Tags) == 0 {
		return ""
	}
	parts := make([]string, len(t.Tags))
	for i, tag := range t.Tags {
		fg := paletteFor(tag).Fg
		if t.Status == task.StatusDone {
			fg = dimmedTagPalette.Fg
		}
		parts[i] = lipgloss.NewStyle().Foreground(fg).Render("+" + tag)
	}
	line := strings.Join(parts, " ")
	if lipgloss.Width(line) > width {
		line = ansi.Truncate(line, width-1, "") + FooterStyle.Render("…")
	}
	return line
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
	switch d := daysBetween(now, due); {
	case done:
		return FooterStyle
	case d <= 0:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(hexPink))
	case d < 3:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(hexAmber))
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
