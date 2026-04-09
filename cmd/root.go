package cmd

import (
	"os"

	"github.com/ferdikt/taskmd-cli/internal/service"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	file string
}

var (
	opts = &rootOptions{}
	svc  = service.New()
)

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "taskmd",
		Short:         "Agent-friendly Markdown task tracker",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.file, "file", "", "Override Task.md path")

	cmd.AddCommand(
		newInitCmd(),
		newListCmd(),
		newShowCmd(),
		newAddCmd(),
		newEditCmd(),
		newStartCmd(),
		newDoneCmd(),
		newReopenCmd(),
		newRemoveCmd(),
		newBulkCmd(),
		newNextCmd(),
		newFmtCmd(),
		newValidateCmd(),
		newVersionCmd(),
	)
	return cmd
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
