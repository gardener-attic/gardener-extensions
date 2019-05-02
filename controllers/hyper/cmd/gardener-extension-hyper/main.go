package main

import (
	"github.com/gardener/gardener-extensions/controllers/hyper/cmd/gardener-extension-hyper/app"
	"github.com/gardener/gardener-extensions/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	controllercmd "github.com/gardener/gardener-extensions/pkg/controller/cmd"
)

func main() {
	log.SetLogger(log.ZapLogger(false))
	cmd := app.NewHyperCommand(controller.SetupSignalHandlerContext())

	if err := cmd.Execute(); err != nil {
		controllercmd.LogErrAndExit(err, "error executing the main command")
	}
}
