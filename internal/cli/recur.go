package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/task"
)

// NewRecurCommand returns "kl recur <id> <daily|weekly|monthly|none>".
func NewRecurCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "recur <id> <daily|weekly|monthly|none>",
		Short: "Set or clear a task's recurrence",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}
			r, err := task.ParseRecurrence(args[1])
			if err != nil {
				return err
			}
			t, err := svc.SetRecurrence(cmd.Context(), id, r)
			if err != nil {
				return err
			}
			if t.Recur == task.RecurNone {
				fmt.Fprintf(cmd.OutOrStdout(), "Recurrence cleared: #%d\n", t.ID)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Recurrence: #%d %s\n", t.ID, t.Recur)
			}
			return nil
		},
	}
}
