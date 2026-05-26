package cli_test

import (
	"strings"
	"testing"

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
