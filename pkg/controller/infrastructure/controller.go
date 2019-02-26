package infrastructure

import (
	"context"

	controllerutil "github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/util"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	// FinalizerName is the infrastructure controller finalizer.
	FinalizerName = "extensions.gardener.cloud/infrastructure"
	name          = "infrastructure-controller"
)

// NewReconciler creates a new reconcile.Reconciler that reconciles
// infrastructure resources of Gardener's `extensions.gardener.cloud` API group.
func NewReconciler(actuator Actuator) reconcile.Reconciler {
	return &reconciler{logger: log.Log.WithName(name), actuator: actuator}
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
	// Predicates are the predicates to use.
	// If unset, GenerationChangedPredicate will be used.
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

// reconciler reconciles Infrastructure resources of Gardener's
// extensions.gardener.cloud` API group.
type reconciler struct {
	logger   logr.Logger
	actuator Actuator

	ctx    context.Context
	client client.Client
}

// InjectFunc enables dependency injection into the actuator.
func (r *reconciler) InjectFunc(f inject.Func) error {
	return f(r.actuator)
}

// InjectClient injects the controller runtime client into the reconciler.
func (r *reconciler) InjectClient(client client.Client) error {
	r.client = client
	return nil
}

// InjectStopChannel is an implementation for getting the respective stop channel managed by the controller-runtime.
func (r *reconciler) InjectStopChannel(stopCh <-chan struct{}) error {
	r.ctx = util.ContextFromStopChannel(stopCh)
	return nil
}

func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	infrastructure := &extensionsv1alpha1.Infrastructure{}
	err := r.client.Get(r.ctx, request.NamespacedName, infrastructure)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if infrastructure.DeletionTimestamp != nil {
		return r.delete(r.ctx, infrastructure)
	}
	return r.reconcile(r.ctx, infrastructure)
}

func (r *reconciler) reconcile(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure) (reconcile.Result, error) {
	if err := controllerutil.EnsureFinalizer(ctx, r.client, FinalizerName, infrastructure); err != nil {
		return reconcile.Result{}, err
	}

	exist, err := r.actuator.Exists(ctx, infrastructure)
	if err != nil {
		return reconcile.Result{}, err
	}

	if exist {
		r.logger.Info("Reconciling infrastructure triggers idempotent update.", "infrastructure", infrastructure.Name)
		if err := r.actuator.Update(ctx, infrastructure); err != nil {
			return controllerutil.ReconcileErr(err)
		}
		return reconcile.Result{}, nil
	}

	r.logger.Info("Reconciling infrastructure triggers idempotent create.", "infrastructure", infrastructure.Name)
	if err := r.actuator.Create(ctx, infrastructure); err != nil {
		r.logger.Error(err, "Unable to create infrastructure", "infrastructure", infrastructure.Name)
		return controllerutil.ReconcileErr(err)
	}
	return reconcile.Result{}, nil
}

func (r *reconciler) delete(ctx context.Context, infrastructure *extensionsv1alpha1.Infrastructure) (reconcile.Result, error) {
	hasFinalizer, err := controllerutil.HasFinalizer(infrastructure, FinalizerName)
	if err != nil {
		r.logger.Error(err, "Could not instantiate finalizer deletion")
		return reconcile.Result{}, err
	}

	if !hasFinalizer {
		r.logger.Info("Reconciling infrastructure causes a no-op as there is no finalizer.", "infrastructure", infrastructure.Name)
		return reconcile.Result{}, nil
	}

	if err := r.actuator.Delete(r.ctx, infrastructure); err != nil {
		r.logger.Error(err, "Error deleting infrastructure", "infrastructure", infrastructure.Name)
		return reconcile.Result{}, err
	}

	r.logger.Info("Infrastructure deletion successful, removing finalizer.", "infrastructure", infrastructure.Name)
	if err := controllerutil.DeleteFinalizer(ctx, r.client, FinalizerName, infrastructure); err != nil {
		r.logger.Error(err, "Error removing finalizer from Infrastructure", "infrastructure", infrastructure.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
