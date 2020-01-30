package app

import (
	"context"

	"github.com/spf13/cobra"
)

// NewHyperCommand creates a new Hyper command consisting of all controllers under this repository.
func NewHyperCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "gardener-extension-hyper",
	}

	cmd.AddCommand()

	return cmd
}
