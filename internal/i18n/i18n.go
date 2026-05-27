// Package i18n provides a tiny message catalog for localizing komlist's
// runtime output. The language is chosen once at startup from KOMLIST_LANG
// (en default, fr supported); unknown keys fall back to English.
//
// Only komlist's own strings are localized. Cobra's structural help words
// ("Usage:", "Flags:", …) remain English.
package i18n

import (
	"fmt"
	"strings"
)

// Key identifies a message in the catalog.
type Key string

// Message keys.
const (
	KeyCreated      Key = "created"
	KeyUpdated      Key = "updated"
	KeyRenamed      Key = "renamed"
	KeyDeleted      Key = "deleted"
	KeyTagged       Key = "tagged"
	KeyTaggedNone   Key = "tagged_none"
	KeyTagRemoved   Key = "tag_removed"
	KeyTagRenamed   Key = "tag_renamed"
	KeyPriority     Key = "priority"
	KeyDue          Key = "due"
	KeyDueCleared   Key = "due_cleared"
	KeyRecur        Key = "recur"
	KeyRecurCleared Key = "recur_cleared"
	KeyNoteAdded    Key = "note_added"
	KeyNotesCleared Key = "notes_cleared"
	KeyMissingNote  Key = "missing_note"
	KeyBlocked      Key = "blocked"
	KeyUnblocked    Key = "unblocked"
	KeyNoTasks      Key = "no_tasks"
	KeyNoTags       Key = "no_tags"
	KeyBlockedBy    Key = "blocked_by"

	// Board footer words.
	KeyStatDone     Key = "stat_done"
	KeyStatDoing    Key = "stat_doing"
	KeyStatBlocked  Key = "stat_blocked"
	KeyStatTodo     Key = "stat_todo"
	KeyStatComplete Key = "stat_complete"

	// Column / field labels.
	KeyColID         Key = "col_id"
	KeyColStatus     Key = "col_status"
	KeyColPrio       Key = "col_prio"
	KeyColTags       Key = "col_tags"
	KeyColDue        Key = "col_due"
	KeyColTitle      Key = "col_title"
	KeyColUpdated    Key = "col_updated"
	KeyColTag        Key = "col_tag"
	KeyColCount      Key = "col_count"
	KeyFieldTitle    Key = "field_title"
	KeyFieldStatus   Key = "field_status"
	KeyFieldPriority Key = "field_priority"
	KeyFieldTags     Key = "field_tags"
	KeyFieldDue      Key = "field_due"
	KeyFieldRecur    Key = "field_recur"
	KeyFieldDepends  Key = "field_depends"
	KeyFieldCreated  Key = "field_created"
	KeyFieldUpdated  Key = "field_updated"
	KeyNotesHeader   Key = "notes_header"

	// Error messages (mapped from service/storage sentinels at print time).
	KeyErrPrefix          Key = "err_prefix"
	KeyErrNotFound        Key = "err_not_found"
	KeyErrEmptyTitle      Key = "err_empty_title"
	KeyErrEmptyTag        Key = "err_empty_tag"
	KeyErrEmptyNote       Key = "err_empty_note"
	KeyErrSelfDependency  Key = "err_self_dependency"
	KeyErrDependencyCycle Key = "err_dependency_cycle"
)

var catalog = map[string]map[Key]string{
	"en": {
		KeyCreated:            "Created: #%d [%s] %s",
		KeyUpdated:            "Updated: #%d [%s] %s",
		KeyRenamed:            "Renamed: #%d [%s] %s",
		KeyDeleted:            "Deleted: #%d",
		KeyTagged:             "Tagged: #%d [%s]",
		KeyTaggedNone:         "Tagged: #%d (none)",
		KeyTagRemoved:         "Removed tag %q from %d task(s).",
		KeyTagRenamed:         "Renamed tag %q -> %q on %d task(s).",
		KeyPriority:           "Priority: #%d [%s]",
		KeyDue:                "Due: #%d %s",
		KeyDueCleared:         "Due: #%d (cleared)",
		KeyRecur:              "Recurrence: #%d %s",
		KeyRecurCleared:       "Recurrence cleared: #%d",
		KeyNoteAdded:          "Note added: #%d (%d total)",
		KeyNotesCleared:       "Notes cleared: #%d",
		KeyMissingNote:        "missing note text (or pass --clear)",
		KeyBlocked:            "Blocked: #%d now depends on #%d",
		KeyUnblocked:          "Unblocked: #%d no longer depends on #%d",
		KeyNoTasks:            "No tasks.",
		KeyNoTags:             "No tags.",
		KeyBlockedBy:          "Blocked by: %s",
		KeyStatDone:           "%d done",
		KeyStatDoing:          "%d doing",
		KeyStatBlocked:        "%d blocked",
		KeyStatTodo:           "%d todo",
		KeyStatComplete:       "%d%% complete",
		KeyColID:              "ID",
		KeyColStatus:          "STATUS",
		KeyColPrio:            "PRIO",
		KeyColTags:            "TAGS",
		KeyColDue:             "DUE",
		KeyColTitle:           "TITLE",
		KeyColUpdated:         "UPDATED",
		KeyColTag:             "TAG",
		KeyColCount:           "COUNT",
		KeyFieldTitle:         "Title",
		KeyFieldStatus:        "Status",
		KeyFieldPriority:      "Priority",
		KeyFieldTags:          "Tags",
		KeyFieldDue:           "Due",
		KeyFieldRecur:         "Recur",
		KeyFieldDepends:       "Depends on",
		KeyFieldCreated:       "Created",
		KeyFieldUpdated:       "Updated",
		KeyNotesHeader:        "Notes:",
		KeyErrPrefix:          "Error:",
		KeyErrNotFound:        "task not found",
		KeyErrEmptyTitle:      "title must not be empty",
		KeyErrEmptyTag:        "tag must not be empty",
		KeyErrEmptyNote:       "note must not be empty",
		KeyErrSelfDependency:  "a task cannot depend on itself",
		KeyErrDependencyCycle: "dependency would create a cycle",
	},
	"fr": {
		KeyCreated:            "Créée : #%d [%s] %s",
		KeyUpdated:            "Mise à jour : #%d [%s] %s",
		KeyRenamed:            "Renommée : #%d [%s] %s",
		KeyDeleted:            "Supprimée : #%d",
		KeyTagged:             "Étiquetée : #%d [%s]",
		KeyTaggedNone:         "Étiquetée : #%d (aucune)",
		KeyTagRemoved:         "Étiquette %q retirée de %d tâche(s).",
		KeyTagRenamed:         "Étiquette %q renommée en %q sur %d tâche(s).",
		KeyPriority:           "Priorité : #%d [%s]",
		KeyDue:                "Échéance : #%d %s",
		KeyDueCleared:         "Échéance : #%d (effacée)",
		KeyRecur:              "Récurrence : #%d %s",
		KeyRecurCleared:       "Récurrence effacée : #%d",
		KeyNoteAdded:          "Note ajoutée : #%d (%d au total)",
		KeyNotesCleared:       "Notes effacées : #%d",
		KeyMissingNote:        "texte de note manquant (ou utilisez --clear)",
		KeyBlocked:            "Bloquée : #%d dépend maintenant de #%d",
		KeyUnblocked:          "Débloquée : #%d ne dépend plus de #%d",
		KeyNoTasks:            "Aucune tâche.",
		KeyNoTags:             "Aucune étiquette.",
		KeyBlockedBy:          "Bloquée par : %s",
		KeyStatDone:           "%d terminée(s)",
		KeyStatDoing:          "%d en cours",
		KeyStatBlocked:        "%d bloquée(s)",
		KeyStatTodo:           "%d à faire",
		KeyStatComplete:       "%d%% terminé",
		KeyColID:              "ID",
		KeyColStatus:          "STATUT",
		KeyColPrio:            "PRIO",
		KeyColTags:            "ÉTIQUETTES",
		KeyColDue:             "ÉCHÉANCE",
		KeyColTitle:           "TITRE",
		KeyColUpdated:         "MAJ",
		KeyColTag:             "ÉTIQUETTE",
		KeyColCount:           "NOMBRE",
		KeyFieldTitle:         "Titre",
		KeyFieldStatus:        "Statut",
		KeyFieldPriority:      "Priorité",
		KeyFieldTags:          "Étiquettes",
		KeyFieldDue:           "Échéance",
		KeyFieldRecur:         "Récurrence",
		KeyFieldDepends:       "Dépend de",
		KeyFieldCreated:       "Créée le",
		KeyFieldUpdated:       "MAJ le",
		KeyNotesHeader:        "Notes :",
		KeyErrPrefix:          "Erreur :",
		KeyErrNotFound:        "tâche introuvable",
		KeyErrEmptyTitle:      "le titre ne doit pas être vide",
		KeyErrEmptyTag:        "l'étiquette ne doit pas être vide",
		KeyErrEmptyNote:       "la note ne doit pas être vide",
		KeyErrSelfDependency:  "une tâche ne peut pas dépendre d'elle-même",
		KeyErrDependencyCycle: "cette dépendance créerait un cycle",
	},
}

var lang = "en"

// Configure sets the active language if supported; otherwise keeps the
// default (en). An empty string leaves the current language unchanged.
func Configure(l string) {
	l = strings.ToLower(strings.TrimSpace(l))
	if l == "" {
		return
	}
	// Accept locale forms like "fr_FR.UTF-8" by taking the leading subtag.
	if i := strings.IndexAny(l, "_.-"); i > 0 {
		l = l[:i]
	}
	if _, ok := catalog[l]; ok {
		lang = l
	}
}

// T returns the localized message for key, formatted with args. Missing keys
// fall back to English, then to the raw key string.
func T(key Key, args ...any) string {
	if tmpl, ok := catalog[lang][key]; ok {
		return format(tmpl, args)
	}
	if tmpl, ok := catalog["en"][key]; ok {
		return format(tmpl, args)
	}
	return string(key)
}

func format(tmpl string, args []any) string {
	if len(args) == 0 {
		return tmpl
	}
	return fmt.Sprintf(tmpl, args...)
}
