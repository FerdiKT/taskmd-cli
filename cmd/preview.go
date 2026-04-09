package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ferdikt/taskmd-cli/internal/preview"
	"github.com/spf13/cobra"
)

func newPreviewCmd() *cobra.Command {
	var port int
	var openBrowser bool

	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Open a local web preview for the current Task.md",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := svc.ResolvePath(mustGetwd(), opts.file)
			if err != nil {
				return err
			}

			url := preview.URL(port)
			server := &http.Server{
				Addr:              fmt.Sprintf("127.0.0.1:%d", port),
				Handler:           preview.New(svc, path).Handler(),
				ReadHeaderTimeout: 5 * time.Second,
			}

			errCh := make(chan error, 1)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errCh <- err
				}
			}()

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "taskmd preview running at %s\n", url)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "watching %s\n", path)
			if openBrowser {
				if err := openURL(url); err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "could not open browser automatically: %v\n", err)
				}
			}

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			defer signal.Stop(sigCh)

			select {
			case sig := <-sigCh:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nreceived %s, shutting down preview\n", sig)
			case err := <-errCh:
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return server.Shutdown(ctx)
		},
	}

	cmd.Flags().IntVar(&port, "port", 4783, "Port for the local preview server")
	cmd.Flags().BoolVar(&openBrowser, "open", true, "Open the preview in your default browser")
	return cmd
}

func openURL(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", url)
	case "linux":
		command = exec.Command("xdg-open", url)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	return command.Start()
}
