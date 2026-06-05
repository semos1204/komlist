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

// Tokyo Night Storm palette — the exact colours from the design spec the
// user shared. Adaptive variants for light terminals approximate the same
// hierarchy with darker text on light backgrounds.
const (
	hexAccent       = "#7aa2f7" // primary blue
	hexAccentLight  = "#3b5fa8"
	hexFg           = "#c0caf5"
	hexFgLight      = "#1a1b26"
	hexFgDim        = "#565f89"
	hexFgDimLight   = "#6b7394"
	hexFgDimmer     = "#414868"
	hexFgDimmest    = "#3b3f56"
	hexGreen        = "#9ece6a"
	hexAmber        = "#e0af68"
	hexPink         = "#f7768e"
	hexCyan         = "#7dcfff"
	hexPillBg       = "#1e2235"
	hexPillBgLight  = "#dde1f5"
	hexSelRowBg     = "#1f2335"
	hexSelRowBgLite = "#e0e6ff"
	hexTabActiveBg  = "#222436"
	hexTabActiveBgL = "#d6dcef"
	hexSep          = "#2f3245"
)

// Accent is the primary highlight colour, used for the TUI title, cursor
// bar, active tab foreground, and tag foreground.
var Accent = lipgloss.AdaptiveColor{Light: hexAccentLight, Dark: hexAccent}

// Shared styles.
var (
	GroupStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: hexFgDimLight, Dark: hexFg})
	IDStyle     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: hexFgDimLight, Dark: hexFgDimmer})
	DoneStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: hexFgDimmer, Dark: hexFgDimmest}).Strikethrough(true)
	FooterStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: hexFgDimLight, Dark: hexFgDim})

	// Pill colours come straight from the design: blue text on a very
	// dark blue background, not grey. Additional palette entries break
	// up the visual sameness when a board has many tags. The cap glyphs
	// use the body bg as their foreground so they "round off" the body's
	// rectangle.
	tagPalettes = []tagPalette{
		// blue
		{
			Bg: lipgloss.AdaptiveColor{Light: hexPillBgLight, Dark: hexPillBg},
			Fg: lipgloss.AdaptiveColor{Light: hexAccentLight, Dark: hexAccent},
		},
		// purple
		{
			Bg: lipgloss.AdaptiveColor{Light: "#ece4f5", Dark: "#231f33"},
			Fg: lipgloss.AdaptiveColor{Light: "#5b3a99", Dark: "#bb9af7"},
		},
		// cyan
		{
			Bg: lipgloss.AdaptiveColor{Light: "#d8eef7", Dark: "#1d272f"},
			Fg: lipgloss.AdaptiveColor{Light: "#2d6f8c", Dark: "#7dcfff"},
		},
		// amber
		{
			Bg: lipgloss.AdaptiveColor{Light: "#f5ead4", Dark: "#2a261d"},
			Fg: lipgloss.AdaptiveColor{Light: "#8c6b1c", Dark: "#e0af68"},
		},
		// green
		{
			Bg: lipgloss.AdaptiveColor{Light: "#e0f0d4", Dark: "#1f2a1e"},
			Fg: lipgloss.AdaptiveColor{Light: "#4e7a2a", Dark: "#9ece6a"},
		},
		// pink (hot)
		{
			Bg: lipgloss.AdaptiveColor{Light: "#f9dde2", Dark: "#2a1d22"},
			Fg: lipgloss.AdaptiveColor{Light: "#a82d4a", Dark: "#f7768e"},
		},
		// orange
		{
			Bg: lipgloss.AdaptiveColor{Light: "#fce8d4", Dark: "#2f221a"},
			Fg: lipgloss.AdaptiveColor{Light: "#a35a1c", Dark: "#ff9e64"},
		},
		// magenta / rose
		{
			Bg: lipgloss.AdaptiveColor{Light: "#fadcf0", Dark: "#2d1d2c"},
			Fg: lipgloss.AdaptiveColor{Light: "#9e3d80", Dark: "#ff79c6"},
		},
	}

	// dimmedTagPalette tones every chip on a "done" row down to a near-
	// invisible neutral, so completed work fades into the background.
	dimmedTagPalette = tagPalette{
		Bg: lipgloss.AdaptiveColor{Light: "#e4e6ef", Dark: "#1c1d2b"},
		Fg: lipgloss.AdaptiveColor{Light: "#7d8198", Dark: hexFgDim},
	}
)

// tagPalette pairs a pill background with its text/cap foreground.
type tagPalette struct {
	Bg lipgloss.AdaptiveColor
	Fg lipgloss.AdaptiveColor
}

// paletteFor picks a stable palette entry from the tag's name via a tiny
// hash, so the same tag always renders in the same colour. Most tags land
// on the blue entry (index 0) thanks to the small palette size.
func paletteFor(tag string) tagPalette {
	var sum uint32
	for _, b := range []byte(tag) {
		sum = sum*31 + uint32(b)
	}
	return tagPalettes[sum%uint32(len(tagPalettes))]
}

// Powerline rounded caps. They require a Nerd Font in the terminal to
// render as half-circles — without one they show up as missing-glyph
// boxes. The visual idea: cap's foreground = body's background, the rest
// of the cap takes whatever background is active around the pill, so the
// cap "rounds off" the body's rectangle.
const (
	pillLeftCap  = ""
	pillRightCap = ""
)

var statusGlyph = map[task.Status]string{
	task.StatusTodo:       "○",
	task.StatusInProgress: "◉", // bullseye — design's "in-progress" mark
	task.StatusBlocked:    "⊘",
	task.StatusDone:       "✓",
}

// StatusColor returns the design's exact colour for each status.
func StatusColor(s task.Status) lipgloss.Color {
	switch s {
	case task.StatusInProgress:
		return lipgloss.Color(hexAccent)
	case task.StatusBlocked:
		return lipgloss.Color(hexPink)
	case task.StatusDone:
		return lipgloss.Color(hexGreen)
	default:
		return lipgloss.Color(hexFgDimmer) // very dim for empty todo circle
	}
}

// PriorityColor returns the design's priority swatch: pink high, amber
// medium, cyan low.
func PriorityColor(p task.Priority) lipgloss.Color {
	switch p {
	case task.PriorityHigh:
		return lipgloss.Color(hexPink)
	case task.PriorityMedium:
		return lipgloss.Color(hexAmber)
	default:
		return lipgloss.Color(hexCyan)
	}
}

// ID renders a task ID right-aligned in a 3-column field so single- and
// double-digit ids line up cleanly.
func ID(id int) string { return IDStyle.Render(fmt.Sprintf("%3d", id)) }

// Bullet renders the coloured status glyph.
func Bullet(s task.Status) string {
	return lipgloss.NewStyle().Foreground(StatusColor(s)).Render(statusGlyph[s])
}

// Priority renders the priority as a small coloured square — the design
// shows it as a 7×7 rounded rect, which a black-square glyph approximates
// best in a terminal. Returns the empty string when no priority is set.
func Priority(p task.Priority) string {
	if p == "" {
		return ""
	}
	return lipgloss.NewStyle().Foreground(PriorityColor(p)).Render("■")
}

// Due renders a due date as a short relative label — "today", "2d",
// "3w" — falling back to "MMM DD" once we're more than a month out (or
// in). Colours follow the design: pink for overdue/today, amber for
// "soon", faint otherwise.
func Due(due time.Time) string {
	now := time.Now()
	label := relativeDue(now, due)
	switch {
	case due.Before(now) && !sameDay(due, now):
		return lipgloss.NewStyle().Foreground(lipgloss.Color(hexPink)).Render(label)
	case sameDay(due, now):
		return lipgloss.NewStyle().Foreground(lipgloss.Color(hexPink)).Render(label)
	case due.Before(now.AddDate(0, 0, 3)):
		return lipgloss.NewStyle().Foreground(lipgloss.Color(hexAmber)).Render(label)
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

// TaskLine renders a task on a single line: "bullet id title prio due recur".
// Used by the CLI board view; the TUI uses the left/right split below to
// right-align tag pills.
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

// TaskLineLeft returns the left-aligned half of a TUI row: bullet, id,
// optional lock, priority dot, and title. The right half (tags + due +
// recur) is rendered separately by TaskLineRight so the TUI can flush it
// against the right edge of the terminal.
func TaskLineLeft(t task.Task, blocked bool) string {
	parts := []string{Bullet(t.Status), ID(t.ID)}
	if blocked {
		parts = append(parts, "\U0001F512")
	}
	if prio := Priority(t.Priority); prio != "" {
		parts = append(parts, prio)
	}
	title := t.Title
	if t.Status == task.StatusDone {
		title = DoneStyle.Render(title)
	}
	parts = append(parts, title)
	return strings.Join(parts, " ")
}

// TaskLineRight returns the right-aligned half of a TUI row: tag pills,
// recurrence marker, and due label, in that visual order. Pills on a
// done task render with the dimmed palette so the row reads as faded.
// Returns the empty string when none of those are set.
func TaskLineRight(t task.Task) string {
	var parts []string
	if pills := TagPills(t.Tags, t.Status == task.StatusDone); pills != "" {
		parts = append(parts, pills)
	}
	if t.Recur != task.RecurNone {
		parts = append(parts, Recur(t.Recur))
	}
	if t.DueAt != nil {
		parts = append(parts, Due(*t.DueAt))
	}
	return strings.Join(parts, " ")
}

// TagPills renders a list of tags as rounded soft-background chips,
// joined by two spaces so they read as separate units. Pass dimmed=true
// for the done-row variant where every chip fades into a near-invisible
// neutral. Otherwise each tag picks a palette entry based on a stable
// hash of its name.
func TagPills(tags []string, dimmed bool) string {
	if len(tags) == 0 {
		return ""
	}
	out := make([]string, len(tags))
	for i, t := range tags {
		palette := paletteFor(t)
		if dimmed {
			palette = dimmedTagPalette
		}
		capStyle := lipgloss.NewStyle().Foreground(palette.Bg)
		bodyStyle := lipgloss.NewStyle().
			Background(palette.Bg).
			Foreground(palette.Fg).
			Padding(0, 1)
		out[i] = capStyle.Render(pillLeftCap) +
			bodyStyle.Render(t) +
			capStyle.Render(pillRightCap)
	}
	return strings.Join(out, "  ")
}

// BgPrefix returns the raw ANSI sequence that sets bg to the given
// adaptive colour for the current colour profile. Useful when callers
// need to keep a row background continuous across nested resets — the
// usual lipgloss wrap won't do that because the inner styles' resets
// kill the outer background. Returns the empty string when bg is nil
// or the profile has no colour.
func BgPrefix(bg lipgloss.TerminalColor) string {
	if bg == nil {
		return ""
	}
	sentinel := "\x00"
	rendered := lipgloss.NewStyle().Background(bg).Render(sentinel)
	idx := strings.Index(rendered, sentinel)
	if idx < 0 {
		return ""
	}
	return rendered[:idx]
}

// WithRowBackground re-applies a row background after every ANSI reset
// embedded in content, so a row stays visually uniform even when its
// inner pieces have their own background or foreground styling. The
// returned string starts with the bg sequence and ends with a final
// reset so the background does not bleed into the next line.
func WithRowBackground(content string, bgSeq string) string {
	if bgSeq == "" {
		return content
	}
	const reset = "\x1b[0m"
	return bgSeq + strings.ReplaceAll(content, reset, reset+bgSeq) + reset
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
