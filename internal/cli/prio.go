package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/task"
)

// NewPriorityCommand returns "kl prio <id> <low|medium|high>".
func NewPriorityCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "prio <id> <low|medium|high>",
		Short: "Set the priority of a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}
			p, err := task.ParsePriority(args[1])
			if err != nil {
				return err
			}
			t, err := svc.SetPriority(cmd.Context(), id, p)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Priority: #%d [%s]\n", t.ID, t.Priority)
			return nil
		},
	}
}
