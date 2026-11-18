package cli_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestCLI_Board(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii) // deterministic output, no escape codes
	_, _, run := newCLI(t)
	run("add", "clean bdd")
	run("add", "migrate API")
	run("add", "buy bread")
	run("tag", "1", "travail")
	run("tag", "2", "travail")
	run("tag", "3", "perso")
	run("status", "2", "in-progress")

	stdout, _, err := run("board")
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	for _, want := range []string{"travail", "perso", "clean bdd", "migrate API", "complete"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("board output missing %q:\n%s", want, stdout)
		}
	}
}

func TestCLI_BoardEmpty(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	_, _, run := newCLI(t)
	stdout, _, err := run("board")
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	if !strings.Contains(stdout, "No tasks.") {
		t.Errorf("expected 'No tasks.', got %q", stdout)
	}
}

func TestCLI_BoardFilterByTag(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	_, _, run := newCLI(t)
	run("add", "a")
	run("add", "b")
	run("tag", "1", "travail")
	run("tag", "2", "perso")

	stdout, _, err := run("board", "travail")
	if err != nil {
		t.Fatalf("board travail: %v", err)
	}
	if !strings.Contains(stdout, "travail") || strings.Contains(stdout, "perso") {
		t.Errorf("board travail should only show the travail group:\n%s", stdout)
	}
}

func TestCLI_BoardDueColumns(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	_, _, run := newCLI(t)
	run("add", "late task")
	run("add", "today task")
	yesterday := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
	today := time.Now().Format(time.DateOnly)
	run("due", "1", yesterday)
	run("due", "2", today)

	stdout, _, err := run("board")
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	for _, want := range []string{"All tasks (2)", "Due today (1)", "Overdue tasks (1)", "due yesterday", "due today"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("board output missing %q:\n%s", want, stdout)
		}
	}
}

func TestCLI_BoardHiddenTags(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	_, _, run := newCLI(t)
	t.Setenv("COLUMNS", "104") // 4 slots of 26: the 3 fixed columns + 1 tag
	for i, tag := range []string{"busy", "busy", "rare"} {
		run("add", "task")
		run("tag", fmt.Sprint(i+1), tag)
	}
	stdout, _, err := run("board")
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	if !strings.Contains(stdout, "busy (2)") {
		t.Errorf("the busiest tag should get the free column:\n%s", stdout)
	}
	if strings.Contains(stdout, "rare (1)") || !strings.Contains(stdout, "+1 tags not shown: rare") {
		t.Errorf("the tag that does not fit should be listed in the footer:\n%s", stdout)
	}
}

func TestCLI_ColumnsRename(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	_, _, run := newCLI(t)
	t.Setenv("COLUMNS", "120")
	run("add", "late")
	run("tag", "1", "finops")
	run("due", "1", time.Now().AddDate(0, 0, -1).Format(time.DateOnly))

	for _, args := range [][]string{
		{"columns", "rename", "all", "Backlog"},
		{"columns", "rename", "today", "Sprint"},
		{"columns", "rename", "overdue", "En feu"},
		{"columns", "rename", "+finops", "FinOps"},
	} {
		if _, _, err := run(args...); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}
	stdout, _, _ := run("board")
	for _, want := range []string{"Backlog (1)", "Sprint (0)", "En feu (1)", "FinOps (1)"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("board missing renamed column %q:\n%s", want, stdout)
		}
	}
	if stdout, _, _ = run("board", "finops"); !strings.Contains(stdout, "FinOps (1)") {
		t.Errorf("a board filtered on a tag should use the tag's custom title:\n%s", stdout)
	}

	list, _, _ := run("columns")
	for _, want := range []string{"all", "Backlog", "+finops", "FinOps"} {
		if !strings.Contains(list, want) {
			t.Errorf("kl columns missing %q:\n%s", want, list)
		}
	}

	run("columns", "reset", "all")
	if stdout, _, _ = run("board"); !strings.Contains(stdout, "All tasks (1)") {
		t.Errorf("reset should restore the default title:\n%s", stdout)
	}
	if _, _, err := run("columns", "rename", "finops", "X"); err == nil {
		t.Error("a tag key without '+' should be rejected")
	}
}
