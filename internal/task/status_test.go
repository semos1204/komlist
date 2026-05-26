package task

import "testing"

func TestStatusValid(t *testing.T) {
	for _, s := range AllStatuses() {
		if !s.Valid() {
			t.Errorf("expected %q to be valid", s)
		}
	}
	if Status("nope").Valid() {
		t.Error("expected \"nope\" to be invalid")
	}
}

func TestParseStatus(t *testing.T) {
	got, err := ParseStatus("todo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != StatusTodo {
		t.Errorf("got %q, want %q", got, StatusTodo)
	}

	if _, err := ParseStatus("nope"); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestAllStatusesLen(t *testing.T) {
	if got := len(AllStatuses()); got != 4 {
		t.Errorf("got %d statuses, want 4", got)
	}
}
