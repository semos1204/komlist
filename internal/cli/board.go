package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/render"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/task"
)

const untaggedGroup = "(untagged)"

// NewBoardCommand returns "kl board [tag]", a colored, taskbook-style view
// grouping tasks by tag and ordering each group by urgency. A positional
// tag argument shows only that board; --status filters by status.
func NewBoardCommand(svc *service.TaskService) *cobra.Command {
	var statusFlag string
	cmd := &cobra.Command{
		Use:   "board [tag]",
		Short: "Show tasks grouped by tag, colored (taskbook-style)",
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
			renderBoard(cmd.OutOrStdout(), tasks, blocked)
			return nil
		},
	}
	cmd.Flags().StringVarP(&statusFlag, "status", "s", "",
		"filter by status (todo, in-progress, blocked, done)")
	return cmd
}

func renderBoard(w io.Writer, tasks []task.Task, blocked map[int]bool) {
	if len(tasks) == 0 {
		fmt.Fprintln(w, i18n.T(i18n.KeyNoTasks))
		return
	}
	groups := groupByTag(tasks)
	for _, name := range sortedGroupNames(groups) {
		fmt.Fprintln(w, " "+render.GroupStyle.Render(name))
		for _, t := range groups[name] {
			fmt.Fprintln(w, "  "+render.TaskLine(t, blocked[t.ID]))
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, renderStats(tasks))
}

// groupByTag buckets tasks by tag, preserving the incoming (urgency-sorted)
// order within each bucket. Multi-tag tasks appear in each of their groups;
// untagged tasks fall under untaggedGroup.
func groupByTag(tasks []task.Task) map[string][]task.Task {
	groups := make(map[string][]task.Task)
	for _, t := range tasks {
		if len(t.Tags) == 0 {
			groups[untaggedGroup] = append(groups[untaggedGroup], t)
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
		case names[i] == untaggedGroup:
			return false
		case names[j] == untaggedGroup:
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
