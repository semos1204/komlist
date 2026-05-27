package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/task"
)

// NewStatusCommand returns "komlist status <id> <new-status>".
func NewStatusCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "status <id> <new-status>",
		Short: "Change the status of a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}
			st, err := task.ParseStatus(args[1])
			if err != nil {
				return err
			}
			t, err := svc.ChangeStatus(cmd.Context(), id, st)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyUpdated, t.ID, t.Status, t.Title))
			return nil
		},
	}
}
