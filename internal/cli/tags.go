package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/service"
)

// NewTagsCommand returns "kl tags" and its sub-commands. Bare `kl tags`
// lists every distinct tag with its task count; `kl tags delete <tag>`
// removes a tag from every task; `kl tags rename <old> <new>` renames a
// tag everywhere.
func NewTagsCommand(svc *service.TaskService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "List or manage tags across all tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tags, err := svc.Tags(cmd.Context())
			if err != nil {
				return err
			}
			if len(tags) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyNoTags))
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyColTag), i18n.T(i18n.KeyColCount))
			for _, tc := range tags {
				fmt.Fprintf(tw, "%s\t%d\n", tc.Tag, tc.Count)
			}
			return tw.Flush()
		},
	}
	cmd.AddCommand(
		newTagsDeleteCommand(svc),
		newTagsRenameCommand(svc),
	)
	return cmd
}

func newTagsDeleteCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <tag>",
		Aliases: []string{"rm"},
		Short:   "Remove a tag from every task that carries it",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := svc.DeleteTag(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyTagRemoved, args[0], n))
			return nil
		},
	}
}

func newTagsRenameCommand(svc *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:     "rename <old> <new>",
		Aliases: []string{"mv"},
		Short:   "Rename a tag across every task",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := svc.RenameTag(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyTagRenamed, args[0], args[1], n))
			return nil
		},
	}
}
