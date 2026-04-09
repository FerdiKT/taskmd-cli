package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <id> [<id>...]",
		Short: "Remove tasks",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			if err := svc.Remove(path, args); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed %d task(s)\n", len(args))
			return err
		},
	}
}
