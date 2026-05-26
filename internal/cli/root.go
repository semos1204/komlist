// Package cli wires the komlist sub-commands. Each NewXxxCommand factory
// receives the dependencies it needs as arguments — no package-level state.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/service"
)

// NewRootCommand returns the root komlist command with all sub-commands
// attached.
func NewRootCommand(svc *service.TaskService) *cobra.Command {
	root := &cobra.Command{
		Use:           "kl",
		Short:         "A small command-line task manager (komlist)",
		Long:          "kl (komlist) tracks lightweight tasks with statuses (todo, in-progress, blocked, done) stored as JSON in ~/.komlist/tasks.json.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		NewAddCommand(svc),
		NewListCommand(svc),
		NewStatusCommand(svc),
		NewEditCommand(svc),
		NewTagCommand(svc),
		NewPriorityCommand(svc),
		NewDueCommand(svc),
		NewDeleteCommand(svc),
	)
	return root
}
