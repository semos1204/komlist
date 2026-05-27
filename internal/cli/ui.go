package cli

import (
	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/tui"
)

// NewUICommand returns "kl ui", an interactive terminal UI over the tasks.
func NewUICommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Interactive terminal UI (j/k to move, space to cycle status, q to quit)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return tui.Run(svc)
		},
	}
}
