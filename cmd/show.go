package cmd

import (
	"github.com/ferdikt/taskmd-cli/internal/output"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a single task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			task, err := svc.Show(path, args[0])
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
