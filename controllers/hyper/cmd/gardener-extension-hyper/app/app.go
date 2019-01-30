package app

import (
	"context"
	coreosalibaba "github.com/gardener/gardener-extensions/controllers/os-coreos-alibaba/cmd/gardener-extension-os-coreos-alibaba/app"
	coreos "github.com/gardener/gardener-extensions/controllers/os-coreos/cmd/gardener-extension-os-coreos/app"
	"github.com/spf13/cobra"
)

// NewHyperCommand creates a new Hyper command consisting of all controllers under this repository.
func NewHyperCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "gardener-extension-hyper",
	}

	cmd.AddCommand(
		coreos.NewControllerCommand(ctx),
		coreosalibaba.NewControllerCommand(ctx),
	)

	return cmd
}
