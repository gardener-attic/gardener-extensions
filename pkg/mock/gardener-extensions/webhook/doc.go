//go:generate mockgen -destination=zz_funcs.go -package=webhook github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/webhook Factory

package webhook

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// Factory allows mocking webhook's Factory functions.
type Factory interface {
	Do(manager.Manager) (webhook.Webhook, error)
}
