package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/service"
)

// NewAddCommand returns "komlist add <title>".
func NewAddCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "add <title>",
		Short: "Create a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := svc.Add(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created: #%d [%s] %s\n", t.ID, t.Status, t.Title)
			return nil
		},
	}
}
