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
