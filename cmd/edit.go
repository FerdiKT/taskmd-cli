package cmd

import (
	"github.com/ferdikt/taskmd-cli/internal/output"
	"github.com/ferdikt/taskmd-cli/internal/service"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	var title string
	var priority string
	var assignee string
	var labels string
	var notes string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit task fields",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			patch := service.EditInput{ID: args[0]}
			if cmd.Flags().Changed("title") {
				patch.Title = &title
			}
			if cmd.Flags().Changed("priority") {
				patch.Priority = &priority
			}
			if cmd.Flags().Changed("assignee") {
				patch.Assignee = &assignee
			}
			if cmd.Flags().Changed("labels") {
				parsed := splitCSV(labels)
				patch.Labels = &parsed
			}
			if cmd.Flags().Changed("notes") {
				patch.Notes = &notes
			}
			task, err := svc.Edit(path, patch)
			if err != nil {
				return err
			}
			if jsonOutput {
				return output.PrintTaskJSON(cmd.OutOrStdout(), task)
			}
			return output.PrintTaskHuman(cmd.OutOrStdout(), task)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Task title")
	cmd.Flags().StringVar(&priority, "priority", "", "Task priority")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Task assignee; pass empty string to clear")
	cmd.Flags().StringVar(&labels, "labels", "", "Comma-separated labels; pass empty string to clear")
	cmd.Flags().StringVar(&notes, "notes", "", "Task notes; pass empty string to clear")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print JSON output")
	return cmd
}
