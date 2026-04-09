package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create docs/Task.md in the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolveInitPath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			if err := svc.Init(path, force); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Initialized %s\n", path)
			return err
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite an existing Task.md")
	return cmd
}
