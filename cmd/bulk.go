package cmd

import (
	"fmt"

	"github.com/ferdikt/taskmd-cli/internal/service"
	"github.com/spf13/cobra"
)

func newBulkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Run bulk task operations from JSON",
	}
	cmd.AddCommand(
		newBulkAddCmd(),
		newBulkEditCmd(),
		newBulkRemoveCmd(),
	)
	return cmd
}

func newBulkAddCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "add --file <path|->",
		Short: "Bulk add tasks from a JSON array",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			data, err := service.ReadInput(file)
			if err != nil {
				return err
			}
			inputs, err := service.DecodeBulkAdd(data)
			if err != nil {
				return err
			}
			if err := svc.BulkAdd(path, inputs); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Added %d task(s)\n", len(inputs))
			return err
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "JSON file path or - for stdin")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func newBulkEditCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "edit --file <path|->",
		Short: "Bulk edit tasks from a JSON array",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			data, err := service.ReadInput(file)
			if err != nil {
				return err
			}
			inputs, err := service.DecodeBulkEdit(data)
			if err != nil {
				return err
			}
			if err := svc.BulkEdit(path, inputs); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Edited %d task(s)\n", len(inputs))
			return err
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "JSON file path or - for stdin")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func newBulkRemoveCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "remove --file <path|->",
		Short: "Bulk remove tasks from a JSON array",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}
			data, err := service.ReadInput(file)
			if err != nil {
				return err
			}
			ids, err := service.DecodeBulkRemove(data)
			if err != nil {
				return err
			}
			if err := svc.BulkRemove(path, ids); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed %d task(s)\n", len(ids))
			return err
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "JSON file path or - for stdin")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}
