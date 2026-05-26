package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/service"
)

// NewNoteCommand returns "kl note <id> [text]". It appends a note to a task,
// or clears every note when --clear is passed.
func NewNoteCommand(svc *service.TaskService) *cobra.Command {
	var clear bool
	cmd := &cobra.Command{
		Use:   "note <id> [text]",
		Short: "Append a note to a task (or clear all with --clear)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}
			if clear {
				t, err := svc.ClearNotes(cmd.Context(), id)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Notes cleared: #%d\n", t.ID)
				return nil
			}
			if len(args) < 2 {
				return fmt.Errorf("missing note text (or pass --clear)")
			}
			t, err := svc.AddNote(cmd.Context(), id, args[1])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Note added: #%d (%d total)\n", t.ID, len(t.Notes))
			return nil
		},
	}
	cmd.Flags().BoolVar(&clear, "clear", false, "remove all notes from the task")
	return cmd
}
