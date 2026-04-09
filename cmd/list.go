package cmd

import (
	"github.com/ferdikt/taskmd-cli/internal/output"
	"github.com/ferdikt/taskmd-cli/internal/service"
	"github.com/ferdikt/taskmd-cli/internal/taskfile"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var statusValue string
	var label string
	var priorityValue string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			filters := service.Filters{Label: label}
			if cmd.Flags().Changed("status") {
				status, err := taskfile.ParseStatus(statusValue)
				if err != nil {
					return err
				}
				filters.Status = status
				filters.HasStatus = true
			}
			if cmd.Flags().Changed("priority") {
				priority, err := taskfile.ParsePriority(priorityValue)
				if err != nil {
					return err
				}
				filters.Priority = priority
				filters.HasPriority = true
			}

			tasks, err := svc.List(path, filters)
			if err != nil {
				return err
			}
			if jsonOutput {
				return output.PrintTasksJSON(cmd.OutOrStdout(), tasks)
			}
			return output.PrintTasksHuman(cmd.OutOrStdout(), tasks)
		},
	}

	cmd.Flags().StringVar(&statusValue, "status", "", "Filter by status")
	cmd.Flags().StringVar(&label, "label", "", "Filter by label")
	cmd.Flags().StringVar(&priorityValue, "priority", "", "Filter by priority")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print JSON output")
	return cmd
}
