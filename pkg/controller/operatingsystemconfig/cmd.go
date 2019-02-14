// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package operatingsystemconfig

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gardener/gardener-extensions/pkg/controller/cmd"
	"github.com/gardener/gardener-extensions/pkg/controller/version"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// DefaultMaxConcurrentReconciles is the default number of maximum concurrent reconciles.
const DefaultMaxConcurrentReconciles = 5

// ExtensionsScheme is the default scheme for extensions, consisting of all Kubernetes built-in
// schemes (client-go/kubernetes/scheme) and the extensions/v1alpha1 scheme.
var ExtensionsScheme = runtime.NewScheme()

func init() {
	utilruntime.Must(scheme.AddToScheme(ExtensionsScheme))
	utilruntime.Must(extensionsv1alpha1.AddToScheme(ExtensionsScheme))
}

// ManagerOptions are options for the creation of a Manager.
type ManagerOptions struct {
	Scheme                  *runtime.Scheme
	LeaderElection          bool
	LeaderElectionID        string
	LeaderElectionNamespace string
	SyncPeriod              *time.Duration
}

// Config produces a ManagerConfig used for instantiating a Manager.
func (m *ManagerOptions) Config() (*ManagerConfig, error) {
	mgrScheme := m.Scheme
	if mgrScheme == nil {
		mgrScheme = ExtensionsScheme
	}

	opts := manager.Options{
		SyncPeriod: m.SyncPeriod,
		Scheme:     mgrScheme,
	}

	opts.LeaderElection = m.LeaderElection
	opts.LeaderElectionID = m.LeaderElectionID
	opts.LeaderElectionNamespace = m.LeaderElectionNamespace

	return &ManagerConfig{Options: opts}, nil
}

// AddFlags adds all ManagerOptions relevant flags to the given FlagSet.
func (m *ManagerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&m.LeaderElection, "leader-election", m.LeaderElection, "Whether to use leader election or not when running this controller manager.")
	fs.StringVar(&m.LeaderElectionID, "leader-election-id", m.LeaderElectionID, "The leader election id to use.")
	fs.StringVar(&m.LeaderElectionNamespace, "leader-election-namespace", m.LeaderElectionNamespace, "The namespace to do leader election in.")
}

// ActuatorArgs are arguments given to the instantiation of an Actuator.
type ActuatorArgs struct {
	Log logr.Logger
}

// ControllerOptions are options used for the creation of a Controller.
type ControllerOptions struct {
	Log                     logr.Logger
	Name                    string
	Type                    string
	Predicates              []predicate.Predicate
	ActuatorFactory         ActuatorFactory
	MaxConcurrentReconciles int
}

// AddFlags adds all ControllerOptions relevant flags to the given FlagSet.
func (c *ControllerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&c.MaxConcurrentReconciles, "max-concurrent-reconciles", c.MaxConcurrentReconciles, "The maximum number of concurrent reconciliations.")
}

// Config produces a ControllerConfig used for instantiating a Controller.
func (c *ControllerOptions) Config() (*ControllerConfig, error) {
	log := c.Log
	if log == nil {
		log = logf.Log
	}
	log = log.WithName(c.Name)

	actuator, err := c.ActuatorFactory(&ActuatorArgs{Log: log.WithName("actuator")})
	if err != nil {
		return nil, err
	}

	predicates := c.Predicates
	if predicates == nil {
		predicates = []predicate.Predicate{GenerationChangedPredicate()}
	}
	predicates = append(predicates, TypePredicate(c.Type))

	return &ControllerConfig{
		Name: c.Name,
		Log:  log.WithName("controller"),
		Options: controller.Options{
			MaxConcurrentReconciles: c.MaxConcurrentReconciles,
			Reconciler:              NewReconciler(log.WithName("reconciler"), actuator),
		},
		Predicates: predicates,
	}, nil
}

// CommandOptions are options used for creating an operating system config controller command.
type CommandOptions struct {
	Manager    *ManagerOptions
	Controller *ControllerOptions
	Mapper     *MapperOptions
}

// Flags yields a NamedFlagSet with all subcomponents relevant for an operating system config
// controller command.
func (c *CommandOptions) Flags() cmd.NamedFlagSet {
	fss := cmd.NamedFlagSet{}

	c.Controller.AddFlags(fss.FlagSet("controller"))

	fs := fss.FlagSet("misc")
	fs.AddGoFlagSet(flag.CommandLine)

	return fss
}

// MapperOptions are options used for creating a secretToOSCMapper.
type MapperOptions struct {
	Type string
}

// NewMapperOptions creates new MapperOptions with the given name.
func NewMapperOptions(typeName string) *MapperOptions {
	return &MapperOptions{
		Type: typeName,
	}
}

// Config returns the MapperConfig for the MapperOptions.
func (m *MapperOptions) Config() (*MapperConfig, error) {
	return &MapperConfig{
		Type: m.Type,
	}, nil
}

// NewManagerOptions creates new ManagerOptions with the given name.
func NewManagerOptions(name string) *ManagerOptions {
	return &ManagerOptions{
		LeaderElectionID:        fmt.Sprintf("%s-leader-election", name),
		LeaderElectionNamespace: v1.NamespaceSystem,
	}
}

// ActuatorFactory is a factory used for creating Actuators.
type ActuatorFactory func(*ActuatorArgs) (Actuator, error)

// NewControllerOptions creates new ControllerOptions with the given name, type name and
// actuator factory.
func NewControllerOptions(name, typeName string, actuatorFactory ActuatorFactory) *ControllerOptions {
	return &ControllerOptions{
		Name:                    name,
		Type:                    typeName,
		ActuatorFactory:         actuatorFactory,
		MaxConcurrentReconciles: DefaultMaxConcurrentReconciles,
	}
}

// NewCommandOptions creates new CommandOptions with the given name, type name and actuator factory.
func NewCommandOptions(name, typeName string, actuatorFactory ActuatorFactory) *CommandOptions {
	return &CommandOptions{
		Manager:    NewManagerOptions(name),
		Controller: NewControllerOptions(name, typeName, actuatorFactory),
		Mapper:     NewMapperOptions(typeName),
	}
}

// Config produces a new CommandConfig used for creating a operating system config command.
func (c *CommandOptions) Config() (*CommandConfig, error) {
	restConfig, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	mgrConfig, err := c.Manager.Config()
	if err != nil {
		return nil, err
	}

	ctrlConfig, err := c.Controller.Config()
	if err != nil {
		return nil, err
	}

	mapperConfig, err := c.Mapper.Config()
	if err != nil {
		return nil, err
	}

	return &CommandConfig{
		REST:       restConfig,
		Manager:    mgrConfig,
		Controller: ctrlConfig,
		Mapper:     mapperConfig,
	}, nil
}

// ManagerConfig is the configuration for creating a operating system config manager.
type ManagerConfig struct {
	Options manager.Options
}

// ControllerConfig is the configuration for creating a operating system config controller.
type ControllerConfig struct {
	Name       string
	Log        logr.Logger
	Predicates []predicate.Predicate
	Options    controller.Options
}

// MapperConfig is the configuration for creating the secretToOSCMapper.
type MapperConfig struct {
	Type string
}

// CommandConfig is the configuration for creating a operating system config command.
type CommandConfig struct {
	REST       *rest.Config
	Manager    *ManagerConfig
	Controller *ControllerConfig
	Mapper     *MapperConfig
}

// Complete fills in any fields not set that are required to have valid data.
func (c *CommandConfig) Complete() *CompletedConfig {
	return &CompletedConfig{&completedConfig{c}}
}

type completedConfig struct {
	*CommandConfig
}

// CompletedConfig is the completed config (all fields set) used to run an operating system
// config command.
type CompletedConfig struct {
	*completedConfig
}

// Run runs the operating system config command with the given completed configuration.
func Run(ctx context.Context, config *CompletedConfig) error {
	log := config.Controller.Log.WithName("entrypoint")
	log.Info("Gardener Controller Extensions", "version", version.Version)

	mgr, err := manager.New(config.REST, config.Manager.Options)
	if err != nil {
		log.Error(err, "Could not instantiate controller-manager")
		return err
	}

	ctrl, err := controller.New(config.Controller.Name, mgr, config.Controller.Options)
	if err != nil {
		log.Error(err, "Could not instantiate controller")
		return err
	}

	if err := ctrl.Watch(&source.Kind{Type: &extensionsv1alpha1.OperatingSystemConfig{}}, &handler.EnqueueRequestForObject{}, config.Controller.Predicates...); err != nil {
		log.Error(err, "Could not watch operating system configs")
		return err
	}

	if err := ctrl.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: SecretToOSCMapper(mgr.GetClient(), config.Mapper.Type)}); err != nil {
		log.Error(err, "Could not watch secrets")
		return err
	}

	return mgr.Start(ctx.Done())
}
