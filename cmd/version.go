package cmd

import (
	"fmt"

	"github.com/ferdikt/taskmd-cli/internal/buildinfo"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\ncommit: %s\ndate: %s\n", buildinfo.Summary(), buildinfo.Commit, buildinfo.Date)
			return err
		},
	}
}
