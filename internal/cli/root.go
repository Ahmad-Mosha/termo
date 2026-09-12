// Package cli wires termo's commands together.
package cli

import (
	"context"
	"os"
	"syscall"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"
)

// Execute runs the termo command line.
func Execute(ctx context.Context) error {
	root := &cobra.Command{
		Use:   "termo",
		Short: "Everyday tools for your terminal",
	}
	return fang.Execute(ctx, root, fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM))
}
