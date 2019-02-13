package infrastructure

import (
	"context"
	"fmt"

	"github.com/gardener/gardener-extensions/pkg/util"

	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"sigs.k8s.io/controller-runtime/pkg/predicate"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/source"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// Finalizer is the infrastructure controller finalizer.
	Finalizer = "extensions.gardener.cloud/infrastructure"
	name      = "infrastructure-controller"
)

// NewReconciler creates a new reconcile.Reconciler that reconciles
// infrastructure resources of Gardener's `extensions.gardener.cloud` API group.
func NewReconciler(actuator Actuator) reconcile.Reconciler {
	return &infrastructureReconciler{logger: log.Log.WithName(name), actuator: actuator}
}

// AddArgs are arguments for adding an infrastructure controller to a manager.
type AddArgs struct {
	// Actuator is an infrastructure actuator.
	Actuator Actuator
	// Type is the infrastructure type the actuator supports.
	Type string
	// ControllerOptions are the controller options used for creating a controller.
	// The options.Reconciler is always overridden with a reconciler created from the
	// given actuator.
	ControllerOptions controller.Options

	Predicates []predicate.Predicate
}

// Add creates a new Infrastructure Controller and adds it to the Manager.
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, args AddArgs) error {
	args.ControllerOptions.Reconciler = NewReconciler(args.Actuator)
	return add(mgr, args.Type, args.ControllerOptions, args.Predicates)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, typeName string, options controller.Options, predicates []predicate.Predicate) error {
	ctrl, err := controller.New(name, mgr, options)
	if err != nil {
		return err
	}

	if predicates == nil {
		predicates = append(predicates, GenerationChangedPredicate())
	}
	predicates = append(predicates, TypePredicate(typeName))

	if err := ctrl.Watch(&source.Kind{Type: &extensionsv1alpha1.Infrastructure{}}, &handler.EnqueueRequestForObject{}, predicates...); err != nil {
		return err
	}
	return nil
}

// infrastructureReconciler reconciles Infrastructure resources of Gardener's
// extensions.gardener.cloud` API group.
type infrastructureReconciler struct {
	logger   logr.Logger
	actuator Actuator

	ctx    context.Context
	client client.Client
}

// InjectFunc enables dependency injection into the actuator.
func (i *infrastructureReconciler) InjectFunc(f inject.Func) error {
	return f(i.actuator)
}

// InjectClient injects the controller runtime client into the reconciler.
func (i *infrastructureReconciler) InjectClient(client client.Client) error {
	i.client = client
	return nil
}

// InjectStopChannel is an implementation for getting the respective stop channel managed by the controller-runtime.
func (i *infrastructureReconciler) InjectStopChannel(stopCh <-chan struct{}) error {
	i.ctx = util.ContextFromStopChannel(stopCh)
	return nil
}

func (i *infrastructureReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	infrastructure := &extensionsv1alpha1.Infrastructure{}
	err := i.client.Get(i.ctx, request.NamespacedName, infrastructure)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	var (
		name                 = infrastructure.Name
		infraType            = infrastructure.Spec.Type
		shootNamespaceInSeed = infrastructure.Namespace

		finalizers = sets.NewString(infrastructure.Finalizers...)
	)

	i.logger.Info(fmt.Sprintf("Reconciling infrastructure of type: %s in namespace: %s", infraType, shootNamespaceInSeed))
	// If the infrastructure resource has not been deleted (no timestamp) and contains no finalizers add one
	if infrastructure.ObjectMeta.DeletionTimestamp.IsZero() && !finalizers.Has(Finalizer) {
		finalizers.Insert(Finalizer)
		infrastructure.Finalizers = finalizers.UnsortedList()
		if err = i.client.Update(i.ctx, infrastructure); err != nil {
			i.logger.Info("Failed to add finalizer to object %v due to err %v", name, err)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if !infrastructure.DeletionTimestamp.IsZero() {
		if !finalizers.Has(Finalizer) {
			return reconcile.Result{}, nil
		}

		if err := i.actuator.Delete(i.ctx, infrastructure); err != nil {
			return reconcile.Result{}, err
		}
		i.logger.Info("Infrastructure deletion went through, removing finalizer!")
		finalizers.Delete(Finalizer)
		infrastructure.Finalizers = finalizers.UnsortedList()

		if err = i.client.Update(i.ctx, infrastructure); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	infraExists, err := i.actuator.Exists(i.ctx, infrastructure)
	if err != nil {
		return reconcile.Result{}, err
	}

	i.logger.Info("reconciling infrastructure object")
	if infraExists {
		if err := i.actuator.Update(i.ctx, infrastructure); err != nil {
			return extensionscontroller.ReconcileErr(err)
		}
	}
	if err := i.actuator.Create(i.ctx, infrastructure); err != nil {
		return extensionscontroller.ReconcileErr(err)
	}

	return reconcile.Result{}, nil
}
