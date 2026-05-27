package cli

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/service"
)

// joinIDs renders a list of task IDs as "#2, #5".
func joinIDs(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("#%d", id)
	}
	return strings.Join(parts, ", ")
}

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
			fmt.Fprintf(tw, "%s\t#%d\n", i18n.T(i18n.KeyColID), t.ID)
			fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyFieldTitle), t.Title)
			fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyFieldStatus), t.Status)
			fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyFieldPriority), fallback(string(t.Priority), "-"))
			fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyFieldTags), fallback(strings.Join(t.Tags, ", "), "-"))
			due := "-"
			if t.DueAt != nil {
				due = t.DueAt.Format(time.DateOnly)
			}
			fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyFieldDue), due)
			fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyFieldRecur), fallback(string(t.Recur), "-"))
			if len(t.DependsOn) > 0 {
				fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyFieldDepends), joinIDs(t.DependsOn))
			}
			fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyFieldCreated), t.CreatedAt.Format("2006-01-02 15:04"))
			fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyFieldUpdated), t.UpdatedAt.Format("2006-01-02 15:04"))
			_ = tw.Flush()

			if inc := svc.IncompleteDeps(cmd.Context(), t); len(inc) > 0 {
				fmt.Fprintf(out, "\n\U0001F512 %s\n", i18n.T(i18n.KeyBlockedBy, joinIDs(inc)))
			}
			if len(t.Notes) > 0 {
				fmt.Fprintln(out, "\n"+i18n.T(i18n.KeyNotesHeader))
				for i, n := range t.Notes {
					fmt.Fprintf(out, "  %d. %s\n", i+1, n)
				}
			}
			return nil
		},
	}
}
