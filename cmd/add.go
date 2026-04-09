package cmd

import (
	"strings"

	"github.com/ferdikt/taskmd-cli/internal/output"
	"github.com/ferdikt/taskmd-cli/internal/service"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var priority string
	var assignee string
	var labels string
	var notes string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			task, err := svc.Add(path, service.AddInput{
				Title:    args[0],
				Priority: priority,
				Assignee: assignee,
				Labels:   splitCSV(labels),
				Notes:    notes,
			})
			if err != nil {
				return err
			}
			if jsonOutput {
				return output.PrintTaskJSON(cmd.OutOrStdout(), task)
			}
			return output.PrintTaskHuman(cmd.OutOrStdout(), task)
		},
	}
	cmd.Flags().StringVar(&priority, "priority", "none", "Task priority")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Task assignee")
	cmd.Flags().StringVar(&labels, "labels", "", "Comma-separated labels")
	cmd.Flags().StringVar(&notes, "notes", "", "Task notes")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print JSON output")
	return cmd
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}
