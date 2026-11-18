package boardcfg

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestValidateKey(t *testing.T) {
	for _, ok := range []string{"all", "today", "overdue", "+finops"} {
		if err := ValidateKey(ok); err != nil {
			t.Errorf("ValidateKey(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"", "finops", "+", "+  ", "All"} {
		if err := ValidateKey(bad); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("ValidateKey(%q) = %v, want ErrInvalidKey", bad, err)
		}
	}
}

func TestFile_RenameResetRoundTrip(t *testing.T) {
	s := NewFile(filepath.Join(t.TempDir(), "sub", "columns.json"))

	got, err := s.Load() // missing file = no custom titles
	if err != nil || len(got) != 0 {
		t.Fatalf("Load on missing file = %v, %v; want empty, nil", got, err)
	}
	if err := Rename(s, "all", "  Backlog  "); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if err := Rename(s, "+finops", "FinOps"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	got, _ = s.Load()
	if got.Title("all", "x") != "Backlog" || got.Title("+finops", "x") != "FinOps" {
		t.Fatalf("titles after rename = %v", got)
	}
	if err := Reset(s, "all"); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	got, _ = s.Load()
	if got.Title("all", "default") != "default" {
		t.Errorf("reset should restore the default, got %v", got)
	}
}

func TestRename_Rejects(t *testing.T) {
	s := NewMemory()
	if err := Rename(s, "all", "   "); !errors.Is(err, ErrEmptyTitle) {
		t.Errorf("blank title: err = %v, want ErrEmptyTitle", err)
	}
	if err := Rename(s, "nope", "X"); !errors.Is(err, ErrInvalidKey) {
		t.Errorf("bad key: err = %v, want ErrInvalidKey", err)
	}
}
