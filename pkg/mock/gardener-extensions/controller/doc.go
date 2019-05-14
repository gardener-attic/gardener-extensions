//go:generate mockgen -destination=zz_funcs.go -package=controller github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/controller AddToManager

package controller

import "sigs.k8s.io/controller-runtime/pkg/manager"

// AddToManager allows mocking controller's AddToManager functions.
type AddToManager interface {
	Do(manager.Manager) error
}
