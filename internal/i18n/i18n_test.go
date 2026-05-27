package i18n

import "testing"

func TestT_EnglishDefault(t *testing.T) {
	Configure("en")
	if got := T(KeyNoTasks); got != "No tasks." {
		t.Errorf("got %q, want %q", got, "No tasks.")
	}
}

func TestT_French(t *testing.T) {
	Configure("fr")
	defer Configure("en")
	if got := T(KeyNoTasks); got != "Aucune tâche." {
		t.Errorf("got %q", got)
	}
	if got := T(KeyDeleted, 3); got != "Supprimée : #3" {
		t.Errorf("got %q", got)
	}
}

func TestConfigure_ParsesLocale(t *testing.T) {
	Configure("fr_FR.UTF-8")
	defer Configure("en")
	if got := T(KeyNoTasks); got != "Aucune tâche." {
		t.Errorf("locale form should select fr, got %q", got)
	}
}

func TestConfigure_UnknownKeepsCurrent(t *testing.T) {
	Configure("en")
	Configure("de") // unsupported → ignored
	if got := T(KeyNoTasks); got != "No tasks." {
		t.Errorf("unknown lang should keep en, got %q", got)
	}
}

func TestT_FallsBackToKey(t *testing.T) {
	Configure("en")
	if got := T(Key("nonexistent")); got != "nonexistent" {
		t.Errorf("missing key should return itself, got %q", got)
	}
}
