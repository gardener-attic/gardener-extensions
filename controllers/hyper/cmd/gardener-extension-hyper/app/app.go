package app

import (
	"context"

	certservice "github.com/gardener/gardener-extensions/controllers/extension-certificate-service/cmd/app"
	coreosalicloud "github.com/gardener/gardener-extensions/controllers/os-coreos-alicloud/cmd/gardener-extension-os-coreos-alicloud/app"
	coreos "github.com/gardener/gardener-extensions/controllers/os-coreos/cmd/gardener-extension-os-coreos/app"
	jeos "github.com/gardener/gardener-extensions/controllers/os-suse-jeos/cmd/gardener-extension-os-suse-jeos/app"
	ubuntu "github.com/gardener/gardener-extensions/controllers/os-ubuntu/cmd/gardener-extension-os-ubuntu/app"
	provideralicloud "github.com/gardener/gardener-extensions/controllers/provider-alicloud/cmd/gardener-extension-provider-alicloud/app"
	provideraws "github.com/gardener/gardener-extensions/controllers/provider-aws/cmd/gardener-extension-provider-aws/app"
	providerazure "github.com/gardener/gardener-extensions/controllers/provider-azure/cmd/gardener-extension-provider-azure/app"
	providergcp "github.com/gardener/gardener-extensions/controllers/provider-gcp/cmd/gardener-extension-provider-gcp/app"
	provideropenstack "github.com/gardener/gardener-extensions/controllers/provider-openstack/cmd/gardener-extension-provider-openstack/app"
	providerpacket "github.com/gardener/gardener-extensions/controllers/provider-packet/cmd/gardener-extension-provider-packet/app"
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
		jeos.NewControllerCommand(ctx),
		ubuntu.NewControllerCommand(ctx),
		provideraws.NewControllerManagerCommand(ctx),
		providerazure.NewControllerManagerCommand(ctx),
		providergcp.NewControllerManagerCommand(ctx),
		provideropenstack.NewControllerManagerCommand(ctx),
		provideralicloud.NewControllerManagerCommand(ctx),
		providerpacket.NewControllerManagerCommand(ctx),
		certservice.NewServiceControllerCommand(ctx),
	)

	return cmd
}
