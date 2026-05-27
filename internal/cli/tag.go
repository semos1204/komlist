package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/service"
)

// NewTagCommand returns "kl tag <id> [tag1,tag2,...]". A missing or empty
// second argument clears all tags.
func NewTagCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "tag <id> [tag1,tag2,...]",
		Short: "Set the tags of a task (omit or pass empty to clear)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}
			var tags []string
			if len(args) == 2 && args[1] != "" {
				tags = strings.Split(args[1], ",")
			}
			t, err := svc.SetTags(cmd.Context(), id, tags)
			if err != nil {
				return err
			}
			if len(t.Tags) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyTaggedNone, t.ID))
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyTagged, t.ID, strings.Join(t.Tags, ",")))
			}
			return nil
		},
	}
}
