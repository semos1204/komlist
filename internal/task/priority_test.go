package task

import "testing"

func TestPriorityValid(t *testing.T) {
	for _, p := range AllPriorities() {
		if !p.Valid() {
			t.Errorf("expected %q to be valid", p)
		}
	}
	if Priority("nope").Valid() {
		t.Error("expected \"nope\" to be invalid")
	}
	if Priority("").Valid() {
		t.Error("expected empty priority to be invalid")
	}
}

func TestParsePriority(t *testing.T) {
	got, err := ParsePriority("high")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != PriorityHigh {
		t.Errorf("got %q, want %q", got, PriorityHigh)
	}
	if _, err := ParsePriority("nope"); err == nil {
		t.Error("expected error for invalid priority")
	}
}

func TestPriorityRank(t *testing.T) {
	if PriorityHigh.Rank() <= PriorityMedium.Rank() {
		t.Error("high should rank above medium")
	}
	if PriorityMedium.Rank() <= PriorityLow.Rank() {
		t.Error("medium should rank above low")
	}
	if Priority("").Rank() != 0 {
		t.Error("unset priority should rank 0")
	}
}
