package infrastructure

import (
	infrainternal "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/infrastructure"
	"github.com/gardener/gardener-extensions/pkg/controller/infrastructure"
)

// NewActuator instantiates a new infrastructure Actuator.
func NewActuator() infrastructure.Actuator {
	return infrainternal.NewActuator()
}
