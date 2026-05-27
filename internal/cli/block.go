package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/service"
)

// NewBlockCommand returns "kl block <id> <blocker-id>": task <id> depends on
// <blocker-id> and stays blocked (sinks in urgency) until the blocker is done.
func NewBlockCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "block <id> <blocker-id>",
		Short: "Make a task depend on another (blocked until the other is done)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, onID, err := twoIDs(args)
			if err != nil {
				return err
			}
			t, err := svc.Block(cmd.Context(), id, onID)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyBlocked, t.ID, onID))
			return nil
		},
	}
}

// NewUnblockCommand returns "kl unblock <id> <blocker-id>".
func NewUnblockCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "unblock <id> <blocker-id>",
		Short: "Remove a dependency from a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, onID, err := twoIDs(args)
			if err != nil {
				return err
			}
			t, err := svc.Unblock(cmd.Context(), id, onID)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyUnblocked, t.ID, onID))
			return nil
		},
	}
}

func twoIDs(args []string) (int, int, error) {
	a, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid id %q: %w", args[0], err)
	}
	b, err := strconv.Atoi(args[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid id %q: %w", args[1], err)
	}
	return a, b, nil
}
