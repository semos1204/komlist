package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/semos1204/komlist/internal/cli"
	"github.com/semos1204/komlist/internal/clock"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/storage"
)

// newCLI returns a fresh service backed by an in-memory repository and a
// fake clock, plus a run() helper that executes one CLI invocation on a
// fresh root command and returns (stdout, stderr, error).
//
// Building a new root per call avoids Cobra flag carry-over between calls
// while keeping shared state on the service so test scenarios stay simple.
func newCLI(t *testing.T) (
	svc *service.TaskService,
	fake *clock.Fake,
	run func(args ...string) (stdout, stderr string, err error),
) {
	t.Helper()
	fake = clock.NewFake(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	svc = service.New(storage.NewMemory(), fake)
	run = func(args ...string) (string, string, error) {
		root := cli.NewRootCommand(svc)
		var out, errb bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&errb)
		root.SetArgs(args)
		err := root.ExecuteContext(context.Background())
		return out.String(), errb.String(), err
	}
	return svc, fake, run
}

func TestCLI_Add(t *testing.T) {
	_, _, run := newCLI(t)
	stdout, _, err := run("add", "first")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	want := "Created: #1 [todo] first"
	if !strings.Contains(stdout, want) {
		t.Errorf("stdout = %q, want substring %q", stdout, want)
	}
}

func TestCLI_AddEmptyTitleErrors(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", ""); err == nil {
		t.Fatal("expected error on empty title, got nil")
	}
}

func TestCLI_ListEmpty(t *testing.T) {
	_, _, run := newCLI(t)
	stdout, _, err := run("list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "No tasks.") {
		t.Errorf("stdout = %q, want %q", stdout, "No tasks.")
	}
}

func TestCLI_AddThenList(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", "first"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := run("add", "second"); err != nil {
		t.Fatalf("add: %v", err)
	}
	stdout, _, err := run("list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, want := range []string{"ID", "STATUS", "TITLE", "first", "second"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout = %q missing %q", stdout, want)
		}
	}
}

func TestCLI_StatusUpdates(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", "task"); err != nil {
		t.Fatalf("add: %v", err)
	}
	stdout, _, err := run("status", "1", "in-progress")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(stdout, "Updated: #1 [in-progress] task") {
		t.Errorf("stdout = %q", stdout)
	}
}

func TestCLI_StatusUnknownIDErrors(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("status", "999", "done"); err == nil {
		t.Fatal("expected error on unknown id, got nil")
	}
}

func TestCLI_StatusInvalidStatusErrors(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", "task"); err != nil {
		t.Fatalf("add: %v", err)
	}
	_, _, err := run("status", "1", "bogus")
	if err == nil {
		t.Fatal("expected error on invalid status, got nil")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("error = %v, want to mention 'invalid status'", err)
	}
}

func TestCLI_Delete(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", "task"); err != nil {
		t.Fatalf("add: %v", err)
	}
	stdout, _, err := run("delete", "1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !strings.Contains(stdout, "Deleted: #1") {
		t.Errorf("stdout = %q", stdout)
	}
	if _, _, err := run("delete", "1"); err == nil {
		t.Fatal("expected error on second delete, got nil")
	}
}

func TestCLI_ListFilter(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", "a"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := run("add", "b"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := run("status", "2", "done"); err != nil {
		t.Fatalf("status: %v", err)
	}
	stdout, _, err := run("list", "--status", "done")
	if err != nil {
		t.Fatalf("list --status: %v", err)
	}
	if !strings.Contains(stdout, "b") || strings.Contains(stdout, "\na ") {
		t.Errorf("filtered list wrong: %q", stdout)
	}
}

func TestCLI_DeleteInvalidIDErrors(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("delete", "abc"); err == nil {
		t.Fatal("expected error on non-numeric id, got nil")
	}
}

func TestCLI_Edit(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", "old"); err != nil {
		t.Fatalf("add: %v", err)
	}
	stdout, _, err := run("edit", "1", "new title")
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if !strings.Contains(stdout, "Renamed: #1") || !strings.Contains(stdout, "new title") {
		t.Errorf("stdout = %q", stdout)
	}
}

func TestCLI_TagSetAndClear(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", "x"); err != nil {
		t.Fatalf("add: %v", err)
	}
	stdout, _, err := run("tag", "1", "work,urgent")
	if err != nil {
		t.Fatalf("tag: %v", err)
	}
	if !strings.Contains(stdout, "work,urgent") {
		t.Errorf("set stdout = %q", stdout)
	}
	stdout, _, err = run("tag", "1")
	if err != nil {
		t.Fatalf("clear: %v", err)
	}
	if !strings.Contains(stdout, "(none)") {
		t.Errorf("clear stdout = %q", stdout)
	}
}

func TestCLI_Priority(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", "x"); err != nil {
		t.Fatalf("add: %v", err)
	}
	stdout, _, err := run("prio", "1", "high")
	if err != nil {
		t.Fatalf("prio: %v", err)
	}
	if !strings.Contains(stdout, "[high]") {
		t.Errorf("stdout = %q", stdout)
	}
	if _, _, err := run("prio", "1", "bogus"); err == nil {
		t.Fatal("expected error on invalid priority")
	}
}

func TestCLI_DueSetAndClear(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", "x"); err != nil {
		t.Fatalf("add: %v", err)
	}
	stdout, _, err := run("due", "1", "2026-06-30")
	if err != nil {
		t.Fatalf("due set: %v", err)
	}
	if !strings.Contains(stdout, "2026-06-30") {
		t.Errorf("set stdout = %q", stdout)
	}
	stdout, _, err = run("due", "1", "none")
	if err != nil {
		t.Fatalf("due clear: %v", err)
	}
	if !strings.Contains(stdout, "(cleared)") {
		t.Errorf("clear stdout = %q", stdout)
	}
	if _, _, err := run("due", "1", "not-a-date"); err == nil {
		t.Fatal("expected error on bad date")
	}
}

func TestCLI_ListAdaptiveColumns(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("add", "x"); err != nil {
		t.Fatalf("add: %v", err)
	}
	// no extra columns when no task has them
	stdout, _, err := run("list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if strings.Contains(stdout, "PRIO") || strings.Contains(stdout, "TAGS") || strings.Contains(stdout, "DUE") {
		t.Errorf("expected compact list, got %q", stdout)
	}
	// set priority -> PRIO column appears
	if _, _, err := run("prio", "1", "high"); err != nil {
		t.Fatalf("prio: %v", err)
	}
	stdout, _, err = run("list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "PRIO") {
		t.Errorf("expected PRIO column, got %q", stdout)
	}
}

func TestCLI_ListSortByPriority(t *testing.T) {
	_, _, run := newCLI(t)
	for _, title := range []string{"low", "high", "med"} {
		if _, _, err := run("add", title); err != nil {
			t.Fatalf("add %s: %v", title, err)
		}
	}
	_, _, _ = run("prio", "1", "low")
	_, _, _ = run("prio", "2", "high")
	_, _, _ = run("prio", "3", "medium")
	stdout, _, err := run("list", "--sort", "priority")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	// "high" row should come before "low" row in the output
	hi := strings.Index(stdout, "high")
	lo := strings.Index(stdout, "low")
	if hi < 0 || lo < 0 || hi > lo {
		t.Errorf("expected high before low, got %q", stdout)
	}
}

func TestCLI_ListInvalidSortErrors(t *testing.T) {
	_, _, run := newCLI(t)
	if _, _, err := run("list", "--sort", "bogus"); err == nil {
		t.Fatal("expected error on invalid sort")
	}
}
