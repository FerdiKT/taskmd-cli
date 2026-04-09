package cmd

import (
	"fmt"

	"github.com/ferdikt/taskmd-cli/internal/taskfile"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	return newStatusCmd("start", "Move tasks to in progress", taskfile.StatusInProgress)
}

func newDoneCmd() *cobra.Command {
	return newStatusCmd("done", "Mark tasks as done", taskfile.StatusDone)
}

func newReopenCmd() *cobra.Command {
	return newStatusCmd("reopen", "Move tasks back to todo", taskfile.StatusTodo)
}

func newStatusCmd(use, short string, status taskfile.Status) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <id> [<id>...]",
		Short: short,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			if err := svc.SetStatus(path, status, args); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Updated %d task(s) to %s\n", len(args), status)
			return err
		},
	}
}
