// Package cli wires the komlist sub-commands. Each NewXxxCommand factory
// receives the dependencies it needs as arguments — no package-level state.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/service"
)

// NewRootCommand returns the root komlist command with all sub-commands
// attached.
//
// Error UX: usage is printed on argument-validation errors (the user typed
// the command wrong) but suppressed on runtime errors raised from RunE
// (e.g. "task not found") — those are shown as plain one-liners by main.
// This is achieved by flipping SilenceUsage to true from PersistentPreRunE,
// which runs only after argument validation has passed.
func NewRootCommand(svc *service.TaskService) *cobra.Command {
	root := &cobra.Command{
		Use:           "kl",
		Short:         "A small command-line task manager (komlist)",
		Long:          "kl (komlist) tracks lightweight tasks with statuses (todo, in-progress, blocked, done) stored as JSON in ~/.komlist/tasks.json.",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			return nil
		},
	}
	root.AddCommand(
		NewAddCommand(svc),
		NewListCommand(svc),
		NewBoardCommand(svc),
		NewUICommand(svc),
		NewShowCommand(svc),
		NewStatusCommand(svc),
		NewEditCommand(svc),
		NewTagCommand(svc),
		NewTagsCommand(svc),
		NewPriorityCommand(svc),
		NewDueCommand(svc),
		NewRecurCommand(svc),
		NewNoteCommand(svc),
		NewBlockCommand(svc),
		NewUnblockCommand(svc),
		NewDeleteCommand(svc),
	)
	return root
}
