package task

import (
	"testing"
	"time"
)

func TestParseRecurrence(t *testing.T) {
	cases := map[string]Recurrence{
		"":        RecurNone,
		"none":    RecurNone,
		"daily":   RecurDaily,
		"weekly":  RecurWeekly,
		"monthly": RecurMonthly,
	}
	for in, want := range cases {
		got, err := ParseRecurrence(in)
		if err != nil {
			t.Errorf("ParseRecurrence(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseRecurrence(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := ParseRecurrence("yearly"); err == nil {
		t.Error("expected error for unknown cadence")
	}
}

func TestParseRecurrence_Intervals(t *testing.T) {
	valid := []string{"2w", "3d", "1mo", "10d", "52w"}
	for _, s := range valid {
		if _, err := ParseRecurrence(s); err != nil {
			t.Errorf("ParseRecurrence(%q) unexpected error: %v", s, err)
		}
	}
	invalid := []string{"0d", "w", "2", "2y", "-1d", "2 w", "1mob"}
	for _, s := range invalid {
		if _, err := ParseRecurrence(s); err == nil {
			t.Errorf("ParseRecurrence(%q) expected error", s)
		}
	}
}

func TestRecurrenceNext_Intervals(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := map[Recurrence]time.Time{
		"3d":  base.AddDate(0, 0, 3),
		"2w":  base.AddDate(0, 0, 14),
		"1mo": base.AddDate(0, 1, 0),
		"6mo": base.AddDate(0, 6, 0),
	}
	for r, want := range cases {
		if got := r.Next(base); !got.Equal(want) {
			t.Errorf("%q.Next = %v, want %v", r, got, want)
		}
	}
}

func TestRecurrenceNext(t *testing.T) {
	base := time.Date(2026, 1, 31, 12, 0, 0, 0, time.UTC)
	if got := RecurDaily.Next(base); !got.Equal(base.AddDate(0, 0, 1)) {
		t.Errorf("daily next = %v", got)
	}
	if got := RecurWeekly.Next(base); !got.Equal(base.AddDate(0, 0, 7)) {
		t.Errorf("weekly next = %v", got)
	}
	if got := RecurMonthly.Next(base); !got.Equal(base.AddDate(0, 1, 0)) {
		t.Errorf("monthly next = %v", got)
	}
	if got := RecurNone.Next(base); !got.Equal(base) {
		t.Errorf("none next should be unchanged, got %v", got)
	}
}
