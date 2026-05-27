package cli

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/task"
)

// NewListCommand returns "kl list [--status x] [--tag x] [--sort x] [--wide]".
// Optional columns (priority, tags, due) only appear when at least one task
// in the result set uses them, keeping default output compact. Pass --wide
// to force every column, even if no task has the value.
func NewListCommand(svc *service.TaskService) *cobra.Command {
	var statusFlag, tagFlag, sortFlag string
	var wideFlag bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks, optionally filtered or sorted",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			filter := service.ListFilter{Tag: tagFlag}
			if statusFlag != "" {
				st, err := task.ParseStatus(statusFlag)
				if err != nil {
					return err
				}
				filter.Status = &st
			}
			sb, err := service.ParseSortBy(sortFlag)
			if err != nil {
				return err
			}
			filter.Sort = sb

			tasks, err := svc.List(cmd.Context(), filter)
			if err != nil {
				return err
			}
			if len(tasks) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyNoTasks))
				return nil
			}
			renderTable(cmd.OutOrStdout(), tasks, wideFlag)
			return nil
		},
	}
	cmd.Flags().StringVarP(&statusFlag, "status", "s", "",
		"filter by status (todo, in-progress, blocked, done)")
	cmd.Flags().StringVarP(&tagFlag, "tag", "t", "",
		"filter by tag (exact match)")
	cmd.Flags().StringVar(&sortFlag, "sort", "",
		"sort by: id (default), due, priority, status")
	cmd.Flags().BoolVarP(&wideFlag, "wide", "w", false,
		"always show every column (PRIO, TAGS, DUE), even when empty")
	return cmd
}

func renderTable(w io.Writer, tasks []task.Task, wide bool) {
	showPrio := wide || anyHas(tasks, func(t task.Task) bool { return t.Priority != "" })
	showTags := wide || anyHas(tasks, func(t task.Task) bool { return len(t.Tags) > 0 })
	showDue := wide || anyHas(tasks, func(t task.Task) bool { return t.DueAt != nil })

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	headers := []string{i18n.T(i18n.KeyColID), i18n.T(i18n.KeyColStatus)}
	if showPrio {
		headers = append(headers, i18n.T(i18n.KeyColPrio))
	}
	if showTags {
		headers = append(headers, i18n.T(i18n.KeyColTags))
	}
	if showDue {
		headers = append(headers, i18n.T(i18n.KeyColDue))
	}
	headers = append(headers, i18n.T(i18n.KeyColTitle), i18n.T(i18n.KeyColUpdated))
	fmt.Fprintln(tw, strings.Join(headers, "\t"))

	for _, t := range tasks {
		cells := []string{fmt.Sprintf("%d", t.ID), string(t.Status)}
		if showPrio {
			cells = append(cells, fallback(string(t.Priority), "-"))
		}
		if showTags {
			cells = append(cells, fallback(strings.Join(t.Tags, ","), "-"))
		}
		if showDue {
			d := "-"
			if t.DueAt != nil {
				d = t.DueAt.Format(time.DateOnly)
			}
			cells = append(cells, d)
		}
		cells = append(cells, t.Title, t.UpdatedAt.Format("2006-01-02 15:04"))
		fmt.Fprintln(tw, strings.Join(cells, "\t"))
	}
	_ = tw.Flush()
}

func anyHas(tasks []task.Task, pred func(task.Task) bool) bool {
	for _, t := range tasks {
		if pred(t) {
			return true
		}
	}
	return false
}

func fallback(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
