package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/render"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/task"
)

// NewBoardCommand returns "kl board [tag]", a colored, topydo-style view of
// bordered columns (all tasks, due today, overdue, then the busiest tags),
// each ordered by urgency and sized to fit the terminal. A positional
// tag argument shows only that board; --status filters by status.
func NewBoardCommand(svc *service.TaskService) *cobra.Command {
	var statusFlag string
	cmd := &cobra.Command{
		Use:   "board [tag]",
		Short: "Show tasks as columns: all, due today, overdue, then tags, sized to the screen",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := service.ListFilter{Sort: service.SortByUrgency}
			if statusFlag != "" {
				st, err := task.ParseStatus(statusFlag)
				if err != nil {
					return err
				}
				filter.Status = &st
			}
			if len(args) == 1 {
				filter.Tag = args[0]
			}
			tasks, err := svc.List(cmd.Context(), filter)
			if err != nil {
				return err
			}
			blocked, err := svc.BlockedSet(cmd.Context())
			if err != nil {
				return err
			}
			renderBoard(cmd.OutOrStdout(), tasks, blocked, filter.Tag)
			return nil
		},
	}
	cmd.Flags().StringVarP(&statusFlag, "status", "s", "",
		"filter by status (todo, in-progress, blocked, done)")
	return cmd
}

func renderBoard(w io.Writer, tasks []task.Task, blocked map[int]bool, tagFilter string) {
	if len(tasks) == 0 {
		fmt.Fprintln(w, i18n.T(i18n.KeyNoTasks))
		return
	}
	now := time.Now()
	width, height := terminalSize(w)
	cols, hidden := boardColumns(tasks, tagFilter, width/render.MinColumnWidth, now)

	footer := []string{renderStats(tasks)}
	if len(hidden) > 0 {
		footer = append(footer, render.FooterStyle.Render(" "+ansi.Truncate(
			i18n.T(i18n.KeyBoardHidden, len(hidden), strings.Join(hidden, ", ")), width-1, "…")))
	}
	if height > 0 {
		height -= len(footer) + 1 // keep the footer and the shell prompt on screen
	}
	fmt.Fprintln(w, render.Board(cols, blocked, width, height, now))
	for _, l := range footer {
		fmt.Fprintln(w, l)
	}
}

// boardColumns builds the topydo-style column set: all tasks, open tasks
// due today, open tasks overdue, then one column per tag — only as many tag
// columns as fit beside the first three in slots columns, busiest tags
// first. The names of the tags left out are returned so the footer can
// list them. When the board is filtered on a tag, the first column carries
// that tag's name and there are no tag columns.
func boardColumns(tasks []task.Task, tagFilter string, slots int, now time.Time) ([]render.Column, []string) {
	var dueToday, overdue []task.Task
	for _, t := range tasks {
		if t.DueAt == nil || t.Status == task.StatusDone {
			continue
		}
		switch {
		case render.IsOverdue(now, *t.DueAt):
			overdue = append(overdue, t)
		case render.IsDueToday(now, *t.DueAt):
			dueToday = append(dueToday, t)
		}
	}
	first := i18n.T(i18n.KeyBoardAll)
	if tagFilter != "" {
		first = tagFilter
	}
	cols := []render.Column{
		{Title: first, Tasks: tasks},
		{Title: i18n.T(i18n.KeyBoardDueToday), Tasks: dueToday},
		{Title: i18n.T(i18n.KeyBoardOverdue), Tasks: overdue},
	}
	if tagFilter != "" {
		return cols, nil
	}
	groups := groupByTag(tasks)
	names := sortedGroupNames(groups)
	sort.SliceStable(names, func(i, j int) bool { return len(groups[names[i]]) > len(groups[names[j]]) })
	shown := min(max(slots-len(cols), 0), len(names))
	for _, name := range names[:shown] {
		cols = append(cols, render.Column{Title: name, Tasks: groups[name]})
	}
	return cols, names[shown:]
}

// terminalSize returns the size of w when it is a terminal, else $COLUMNS
// and $LINES. Width falls back to 120; height falls back to 0 (unlimited),
// so piped output is never cut.
func terminalSize(w io.Writer) (width, height int) {
	if f, ok := w.(*os.File); ok {
		if cw, ch, err := term.GetSize(f.Fd()); err == nil && cw > 0 {
			return cw, ch
		}
	}
	width = 120
	if c, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && c > 0 {
		width = c
	}
	if l, err := strconv.Atoi(os.Getenv("LINES")); err == nil && l > 0 {
		height = l
	}
	return width, height
}

// groupByTag buckets tasks by tag, preserving the incoming (urgency-sorted)
// order within each bucket. Multi-tag tasks appear in each of their groups;
// untagged tasks fall under render.UntaggedGroup.
func groupByTag(tasks []task.Task) map[string][]task.Task {
	groups := make(map[string][]task.Task)
	for _, t := range tasks {
		if len(t.Tags) == 0 {
			groups[render.UntaggedGroup] = append(groups[render.UntaggedGroup], t)
			continue
		}
		for _, tag := range t.Tags {
			groups[tag] = append(groups[tag], t)
		}
	}
	return groups
}

func sortedGroupNames(groups map[string][]task.Task) []string {
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		switch {
		case names[i] == render.UntaggedGroup:
			return false
		case names[j] == render.UntaggedGroup:
			return true
		default:
			return names[i] < names[j]
		}
	})
	return names
}

// renderStats counts statuses over the unique task set (not per-group
// duplicates) and renders the footer line.
func renderStats(tasks []task.Task) string {
	var done, doing, blocked, todo int
	for _, t := range tasks {
		switch t.Status {
		case task.StatusDone:
			done++
		case task.StatusInProgress:
			doing++
		case task.StatusBlocked:
			blocked++
		default:
			todo++
		}
	}
	pct := 0
	if total := len(tasks); total > 0 {
		pct = done * 100 / total
	}
	parts := []string{i18n.T(i18n.KeyStatDone, done), i18n.T(i18n.KeyStatDoing, doing)}
	if blocked > 0 {
		parts = append(parts, i18n.T(i18n.KeyStatBlocked, blocked))
	}
	parts = append(parts, i18n.T(i18n.KeyStatTodo, todo))
	return render.FooterStyle.Render(fmt.Sprintf(" %s — %s", strings.Join(parts, " · "), i18n.T(i18n.KeyStatComplete, pct)))
}
