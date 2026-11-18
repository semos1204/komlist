package render

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/semos1204/komlist/internal/task"
)

func TestDaysBetween_DueDateStoredAsUTCMidnight(t *testing.T) {
	paris := time.FixedZone("CET", 3600)
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, paris)
	cases := []struct {
		due  time.Time
		want int
	}{
		{time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC), -1}, // was 0 before the fix
		{time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC), 0},
		{time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC), -10},
	}
	for _, c := range cases {
		if got := daysBetween(now, c.due); got != c.want {
			t.Errorf("daysBetween(%v, %v) = %d, want %d", now, c.due, got, c.want)
		}
	}
}

func TestBoard_FitsWidthAndWraps(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	tk := task.Task{ID: 7, Title: "A rather long title that must wrap inside its column", Status: task.StatusTodo, Tags: []string{"infra"}}
	cols := []Column{{Title: "A", Tasks: []task.Task{tk}}, {Title: "B"}, {Title: "C"}, {Title: "D"}}

	out := Board(cols, nil, 70, now) // room for 2 columns per row
	for _, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > 70 {
			t.Errorf("line wider than terminal (%d > 70): %q", w, line)
		}
	}
	if got := strings.Count(out, "┌"); got != 4 {
		t.Errorf("want 4 boxes, got %d:\n%s", got, out)
	}
	if lines := strings.Split(out, "\n"); !strings.Contains(lines[1], "A (1)") || !strings.Contains(lines[1], "B (0)") {
		t.Errorf("first row should hold columns A and B:\n%s", out)
	}
	for _, want := range []string{"+infra", "must wrap"} {
		if !strings.Contains(out, want) {
			t.Errorf("board missing %q:\n%s", want, out)
		}
	}
}

func TestDueLabel(t *testing.T) {
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	day := func(d int) time.Time { return time.Date(2026, 11, 18+d, 0, 0, 0, 0, time.UTC) }
	cases := map[int]string{0: "due today", 1: "due tomorrow", -1: "due yesterday", 4: "due in 4 days", -3: "due 3 days ago"}
	for d, want := range cases {
		if got := DueLabel(now, day(d)); got != want {
			t.Errorf("DueLabel(+%d) = %q, want %q", d, got, want)
		}
	}
}
