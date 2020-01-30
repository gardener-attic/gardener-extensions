package app

import (
	"context"

	shootdnsservice "github.com/gardener/gardener-extensions/controllers/extension-shoot-dns-service/cmd/app"
	provideralicloud "github.com/gardener/gardener-extensions/controllers/provider-alicloud/cmd/gardener-extension-provider-alicloud/app"
	provideraws "github.com/gardener/gardener-extensions/controllers/provider-aws/cmd/gardener-extension-provider-aws/app"
	validatoraws "github.com/gardener/gardener-extensions/controllers/provider-aws/cmd/gardener-extension-validator-aws/app"
	providerazure "github.com/gardener/gardener-extensions/controllers/provider-azure/cmd/gardener-extension-provider-azure/app"
	providergcp "github.com/gardener/gardener-extensions/controllers/provider-gcp/cmd/gardener-extension-provider-gcp/app"
	provideropenstack "github.com/gardener/gardener-extensions/controllers/provider-openstack/cmd/gardener-extension-provider-openstack/app"
	providerpacket "github.com/gardener/gardener-extensions/controllers/provider-packet/cmd/gardener-extension-provider-packet/app"
	providervsphere "github.com/gardener/gardener-extensions/controllers/provider-vsphere/cmd/gardener-extension-provider-vsphere/app"

	"github.com/spf13/cobra"
)

// NewHyperCommand creates a new Hyper command consisting of all controllers under this repository.
func NewHyperCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "gardener-extension-hyper",
	}

	cmd.AddCommand(
		provideraws.NewControllerManagerCommand(ctx),
		providerazure.NewControllerManagerCommand(ctx),
		providergcp.NewControllerManagerCommand(ctx),
		provideropenstack.NewControllerManagerCommand(ctx),
		provideralicloud.NewControllerManagerCommand(ctx),
		providerpacket.NewControllerManagerCommand(ctx),
		providervsphere.NewControllerManagerCommand(ctx),
		shootdnsservice.NewServiceControllerCommand(ctx),
		validatoraws.NewValidatorCommand(ctx),
	)

	return cmd
}
