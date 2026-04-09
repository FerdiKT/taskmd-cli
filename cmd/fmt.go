package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fmt",
		Short: "Rewrite Task.md into canonical format",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			if err := svc.Format(path); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Formatted %s\n", path)
			return err
		},
	}
}
