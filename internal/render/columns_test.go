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

	out := Board(cols, nil, 60, 0, now) // room for 2 columns per row
	for _, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > 60 {
			t.Errorf("line wider than terminal (%d > 60): %q", w, line)
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

func TestCard_ClampsTitleAndTags(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	due := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	tk := task.Task{
		ID: 31, Status: task.StatusTodo, Priority: task.PriorityHigh, DueAt: &due,
		Title: "[P1] Engagements RI/SP : couvrir 70-80% du socle compute (On-Demand -> 1 an no-upfront)",
		Tags:  []string{"finops", "engagements", "rds", "ec2", "elasticache"},
	}
	card := Card(tk, false, 38, now)
	lines := strings.Split(card, "\n")
	if len(lines) != 4 { // 2 title lines, 1 tag line, 1 due line
		t.Fatalf("card has %d lines, want 4:\n%s", len(lines), card)
	}
	for _, l := range lines {
		if w := lipgloss.Width(l); w > 38 {
			t.Errorf("line wider than the card (%d > 38): %q", w, l)
		}
	}
	if !strings.Contains(lines[1], "…") || !strings.Contains(lines[2], "…") {
		t.Errorf("long title and tag line should end with an ellipsis:\n%s", card)
	}
	if !strings.Contains(lines[0], "(H)") || !strings.Contains(lines[3], "due 166 days ago") {
		t.Errorf("want priority marker on line 1 and the due label on line 4:\n%s", card)
	}
}

func TestBoard_HeightLimitCutsColumns(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	var tasks []task.Task
	for i := 1; i <= 10; i++ {
		tasks = append(tasks, task.Task{ID: i, Title: "task", Status: task.StatusTodo})
	}
	out := Board([]Column{{Title: "All", Tasks: tasks}}, nil, 40, 12, now)
	if h := lipgloss.Height(out); h > 12 {
		t.Errorf("board is %d lines tall, want at most 12:\n%s", h, out)
	}
	if !strings.Contains(out, "… +") || !strings.Contains(out, "more") {
		t.Errorf("cut column should end with '… +N more':\n%s", out)
	}
	if full := Board([]Column{{Title: "All", Tasks: tasks}}, nil, 40, 0, now); strings.Contains(full, "more") {
		t.Errorf("height 0 means unlimited, nothing should be cut:\n%s", full)
	}
}

func TestDueLabel(t *testing.T) {
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	day := func(d int) time.Time { return time.Date(2026, 11, 18+d, 0, 0, 0, 0, time.UTC) }
	cases := map[int]string{0: "due today", 1: "due tomorrow", -1: "due yesterday", 4: "due in 4 days", -3: "due 3 days ago"}
	for d, want := range cases {
		if got := DueLabel(now, day(d)); got != want {
			t.Errorf("DueLabel(%+d) = %q, want %q", d, got, want)
		}
	}
}

func TestColumnHeader_CutsLongTitleKeepsCount(t *testing.T) {
	got := columnHeader("Un titre de colonne beaucoup trop long", 12, 24)
	if w := lipgloss.Width(got); w > 24 {
		t.Errorf("header %q is %d cells, want at most 24", got, w)
	}
	if !strings.HasSuffix(got, "… (12)") {
		t.Errorf("header %q should keep the count after the ellipsis", got)
	}
}
