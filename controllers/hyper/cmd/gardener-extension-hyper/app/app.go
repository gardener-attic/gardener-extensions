package app

import (
	"context"

	provideralicloud "github.com/gardener/gardener-extensions/controllers/provider-alicloud/cmd/gardener-extension-provider-alicloud/app"
	provideraws "github.com/gardener/gardener-extensions/controllers/provider-aws/cmd/gardener-extension-provider-aws/app"
	validatoraws "github.com/gardener/gardener-extensions/controllers/provider-aws/cmd/gardener-extension-validator-aws/app"
	provideropenstack "github.com/gardener/gardener-extensions/controllers/provider-openstack/cmd/gardener-extension-provider-openstack/app"

	"github.com/spf13/cobra"
)

// NewHyperCommand creates a new Hyper command consisting of all controllers under this repository.
func NewHyperCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "gardener-extension-hyper",
	}

	cmd.AddCommand(
		provideraws.NewControllerManagerCommand(ctx),
		provideropenstack.NewControllerManagerCommand(ctx),
		provideralicloud.NewControllerManagerCommand(ctx),
		validatoraws.NewValidatorCommand(ctx),
	)

	return cmd
}
