package cli

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/service"
)

// NewShowCommand returns "kl show <id>", a full-detail view of a single task
// including its notes.
func NewShowCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show full details of a task, including notes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}
			t, err := svc.Get(cmd.Context(), id)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintf(tw, "ID\t#%d\n", t.ID)
			fmt.Fprintf(tw, "Title\t%s\n", t.Title)
			fmt.Fprintf(tw, "Status\t%s\n", t.Status)
			fmt.Fprintf(tw, "Priority\t%s\n", fallback(string(t.Priority), "-"))
			fmt.Fprintf(tw, "Tags\t%s\n", fallback(strings.Join(t.Tags, ", "), "-"))
			due := "-"
			if t.DueAt != nil {
				due = t.DueAt.Format(time.DateOnly)
			}
			fmt.Fprintf(tw, "Due\t%s\n", due)
			fmt.Fprintf(tw, "Recur\t%s\n", fallback(string(t.Recur), "-"))
			fmt.Fprintf(tw, "Created\t%s\n", t.CreatedAt.Format("2006-01-02 15:04"))
			fmt.Fprintf(tw, "Updated\t%s\n", t.UpdatedAt.Format("2006-01-02 15:04"))
			_ = tw.Flush()

			if len(t.Notes) > 0 {
				fmt.Fprintln(out, "\nNotes:")
				for i, n := range t.Notes {
					fmt.Fprintf(out, "  %d. %s\n", i+1, n)
				}
			}
			return nil
		},
	}
}
