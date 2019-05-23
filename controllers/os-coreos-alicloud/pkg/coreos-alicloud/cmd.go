package coreos

import (
	"github.com/gardener/gardener-extensions/pkg/controller/cmd"
	"github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig"
)

// ControllerSwitchOptions are the cmd.SwitchOptions to add all controllers of this provider to a manager.
func ControllerSwitchOptions() *cmd.SwitchOptions {
	return cmd.NewSwitchOptions(
		cmd.Switch(operatingsystemconfig.ControllerName, AddToManager),
	)
}
