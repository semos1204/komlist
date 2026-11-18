package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/semos1204/komlist/internal/boardcfg"
	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/service"
)

// NewColumnsCommand returns "kl columns" and its sub-commands. Bare
// `kl columns` lists every board column (the three fixed ones, then one per
// existing tag) with its current title; `kl columns rename <column> <title>`
// sets a custom title; `kl columns reset <column>` restores the default.
func NewColumnsCommand(svc *service.TaskService, titles boardcfg.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "columns",
		Short: "List or rename the board columns (all, today, overdue, +<tag>)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			t, err := titles.Load()
			if err != nil {
				return err
			}
			tags, err := svc.Tags(cmd.Context())
			if err != nil {
				return err
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(tw, "%s\t%s\n", i18n.T(i18n.KeyColColumn), i18n.T(i18n.KeyColTitle))
			for _, key := range boardcfg.FixedKeys {
				fmt.Fprintf(tw, "%s\t%s\n", key, t.Title(key, defaultColumnTitle(key)))
			}
			for _, tc := range tags {
				key := boardcfg.TagKey(tc.Tag)
				fmt.Fprintf(tw, "%s\t%s\n", key, t.Title(key, tc.Tag))
			}
			return tw.Flush()
		},
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:     "rename <column> <title>",
			Aliases: []string{"mv"},
			Short:   "Give a column a custom title",
			Args:    cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := boardcfg.Rename(titles, args[0], args[1]); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyColumnRenamed, args[0], args[1]))
				return nil
			},
		},
		&cobra.Command{
			Use:   "reset <column>",
			Short: "Restore a column's default title",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := boardcfg.Reset(titles, args[0]); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyColumnReset, args[0]))
				return nil
			},
		},
	)
	return cmd
}

// defaultColumnTitle returns the localized default title of a fixed
// column key.
func defaultColumnTitle(key string) string {
	switch key {
	case boardcfg.KeyToday:
		return i18n.T(i18n.KeyBoardDueToday)
	case boardcfg.KeyOverdue:
		return i18n.T(i18n.KeyBoardOverdue)
	default:
		return i18n.T(i18n.KeyBoardAll)
	}
}
