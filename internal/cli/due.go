package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/service"
)

// NewDueCommand returns "kl due <id> <YYYY-MM-DD|none>". Pass "none" (or
// an empty argument) to clear an existing due date.
func NewDueCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "due <id> <YYYY-MM-DD|none>",
		Short: "Set or clear the due date of a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}
			var due *time.Time
			if args[1] != "" && args[1] != "none" {
				parsed, err := time.Parse(time.DateOnly, args[1])
				if err != nil {
					return fmt.Errorf("invalid date %q (expected YYYY-MM-DD): %w", args[1], err)
				}
				due = &parsed
			}
			t, err := svc.SetDueAt(cmd.Context(), id, due)
			if err != nil {
				return err
			}
			if t.DueAt == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Due: #%d (cleared)\n", t.ID)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Due: #%d %s\n", t.ID, t.DueAt.Format(time.DateOnly))
			}
			return nil
		},
	}
}
