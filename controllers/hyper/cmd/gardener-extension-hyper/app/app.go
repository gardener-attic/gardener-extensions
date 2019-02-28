package app

import (
	"context"

	coreosalicloud "github.com/gardener/gardener-extensions/controllers/os-coreos-alicloud/cmd/gardener-extension-os-coreos-alicloud/app"
	coreos "github.com/gardener/gardener-extensions/controllers/os-coreos/cmd/gardener-extension-os-coreos/app"
	provideraws "github.com/gardener/gardener-extensions/controllers/provider-aws/cmd/gardener-extension-provider-aws/app"
	"github.com/spf13/cobra"
)

// NewHyperCommand creates a new Hyper command consisting of all controllers under this repository.
func NewHyperCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "gardener-extension-hyper",
	}

	cmd.AddCommand(
		coreos.NewControllerCommand(ctx),
		coreosalicloud.NewControllerCommand(ctx),
		provideraws.NewControllerManagerCommand(ctx),
	)

	return cmd
}
