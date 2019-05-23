package coreos

import (
	"github.com/gardener/gardener-extensions/pkg/controller/cmd"
	"github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig"
)

// ControllerSwitchOptions are the cmd.SwitchOptions for the controllers of this provider.
func ControllerSwitchOptions() *cmd.SwitchOptions {
	return cmd.NewSwitchOptions(
		cmd.Switch(operatingsystemconfig.ControllerName, AddToManager),
	)
}
