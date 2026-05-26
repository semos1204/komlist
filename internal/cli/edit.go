package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/service"
)

// NewEditCommand returns "kl edit <id> <new-title>".
func NewEditCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <id> <new-title>",
		Short: "Rename a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}
			t, err := svc.Rename(cmd.Context(), id, args[1])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Renamed: #%d [%s] %s\n", t.ID, t.Status, t.Title)
			return nil
		},
	}
}
