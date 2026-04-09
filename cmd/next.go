package cmd

import (
	"github.com/ferdikt/taskmd-cli/internal/output"
	"github.com/spf13/cobra"
)

func newNextCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Show the next recommended todo task",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			task, err := svc.Next(path)
			if err != nil {
				return err
			}
			if jsonOutput {
				return output.PrintTaskJSON(cmd.OutOrStdout(), task)
			}
			return output.PrintTaskHuman(cmd.OutOrStdout(), task)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print JSON output")
	return cmd
}
