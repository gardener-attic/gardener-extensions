package terraformer

import (
	"github.com/gardener/gardener/pkg/operation/terraformer"
	"github.com/sirupsen/logrus"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Terraformer is an entity that runs Terraform scripts and returns their outputs.
type Terraformer interface {
	// InitializeWith initializes the Terraformer with the given initializer.
	InitializeWith(initializer terraformer.Initializer) Terraformer
	// SetVariablesEnvironment sets the tfVarsEnvironment for this Terraformer.
	SetVariablesEnvironment(tfVarsEnvironment map[string]string) Terraformer

	// Apply tries to apply the operations of this terraformer.
	Apply() error
	// Destroy destroys all resources of this terraformer.
	Destroy() error

	// ConfigExists checks if some config of a previous run exists.
	ConfigExists() (bool, error)
	// GetOutputStateVariables gets the given output state variables.
	GetStateOutputVariables(variables ...string) (map[string]string, error)
}

type terraformerCompat struct {
	base *terraformer.Terraformer
}

// NewForConfig creates a new Terraformer and its dependencies from the given configuration.
func NewForConfig(
	logger logrus.FieldLogger,
	config *rest.Config,
	purpose,
	namespace,
	name,
	image string,
) (Terraformer, error) {
	base, err := terraformer.NewForConfig(logger, config, purpose, namespace, name, image)
	if err != nil {
		return nil, err
	}

	return &terraformerCompat{base}, nil
}

// New takes a <logger>, a <k8sClient>, a string <purpose>, which describes for what the
// Terraformer is used, a <name>, a <namespace> in which the Terraformer will run, and the
// <image> name for the to-be-used Docker image. It returns a Terraformer struct with initialized
// values for the namespace and the names which will be used for all the stored resources like
// ConfigMaps/Secrets.
func New(
	logger logrus.FieldLogger,
	client client.Client,
	coreV1Client corev1client.CoreV1Interface,
	purpose,
	namespace,
	name,
	image string,
) Terraformer {
	return &terraformerCompat{terraformer.New(logger, client, coreV1Client, purpose, namespace, name, image)}
}

// InitializeWith implements Terraformer.
func (t *terraformerCompat) InitializeWith(initializer terraformer.Initializer) Terraformer {
	return &terraformerCompat{t.base.InitializeWith(initializer)}
}

// SetVariablesEnvironment implements Terraformer.
func (t *terraformerCompat) SetVariablesEnvironment(tfVarsEnvironment map[string]string) Terraformer {
	return &terraformerCompat{t.base.SetVariablesEnvironment(tfVarsEnvironment)}
}

// Apply implements Terraformer.
func (t *terraformerCompat) Apply() error {
	return t.base.Apply()
}

// Destroy implements Terraformer.
func (t *terraformerCompat) Destroy() error {
	return t.base.Destroy()
}

// ConfigExists implements Terraformer.
func (t *terraformerCompat) ConfigExists() (bool, error) {
	return t.base.ConfigExists()
}

// GetStateOutputVariables implements Terraformer.
func (t *terraformerCompat) GetStateOutputVariables(variables ...string) (map[string]string, error) {
	return t.base.GetStateOutputVariables(variables...)
}
