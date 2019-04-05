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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// LeaderElectionFlag is the name of the command line flag to specify whether to do leader election or not.
	LeaderElectionFlag = "leader-election"
	// LeaderElectionIDFlag is the name of the command line flag to specify the leader election ID.
	LeaderElectionIDFlag = "leader-election-id"
	// LeaderElectionNamespaceFlag is the name of the command line flag to specify the leader election namespace.
	LeaderElectionNamespaceFlag = "leader-election-namespace"

	// MaxConcurrentReconcilesFlag is the name of the command line flag to specify the maximum number of
	// concurrent reconciliations a controller can do.
	MaxConcurrentReconcilesFlag = "max-concurrent-reconciles"

	// KubeconfigFlag is the name of the command line flag to specify a kubeconfig used to retrieve
	// a rest.Config for a manager.Manager.
	KubeconfigFlag = clientcmd.RecommendedConfigPathFlag
	// MasterURLFlag is the name of the command line flag to specify the master URL override for
	// a rest.Config of a manager.Manager.
	MasterURLFlag = "master"
)

// LeaderElectionNameID returns a leader election ID for the given name.
func LeaderElectionNameID(name string) string {
	return fmt.Sprintf("%s-leader-election", name)
}

// Flagger adds flags to a given FlagSet.
type Flagger interface {
	// AddFlags adds the flags of this Flagger to the given FlagSet.
	AddFlags(*pflag.FlagSet)
}

type prefixedFlagger struct {
	prefix  string
	flagger Flagger
}

// AddFlags implements Flagger.AddFlags.
func (p *prefixedFlagger) AddFlags(fs *pflag.FlagSet) {
	temp := pflag.NewFlagSet("", pflag.ExitOnError)
	p.flagger.AddFlags(temp)
	temp.VisitAll(func(flag *pflag.Flag) {
		flag.Name = fmt.Sprintf("%s%s", p.prefix, flag.Name)
	})
	fs.AddFlagSet(temp)
}

// PrefixFlagger creates a flagger that prefixes all its flags with the given prefix.
func PrefixFlagger(prefix string, flagger Flagger) Flagger {
	return &prefixedFlagger{prefix, flagger}
}

// PrefixOption creates an option that prefixes all its flags with the given prefix.
func PrefixOption(prefix string, option Option) Option {
	return struct {
		Flagger
		Completer
	}{PrefixFlagger(prefix, option), option}
}

// Completer completes some work.
type Completer interface {
	// Complete completes the work, optionally returning an error.
	Complete() error
}

// Option is a Flagger and Completer.
// It sets command line flags and does some work when the flags have been parsed, optionally producing
// an error.
type Option interface {
	Flagger
	Completer
}

// OptionAggregator is a builder that aggregates multiple options.
type OptionAggregator []Option

// NewOptionAggregator instantiates a new OptionAggregator and registers all given options.
func NewOptionAggregator(options ...Option) OptionAggregator {
	var builder OptionAggregator
	builder.Register(options...)
	return builder
}

// Register registers the given options in this OptionAggregator.
func (b *OptionAggregator) Register(options ...Option) {
	*b = append(*b, options...)
}

// AddFlags implements Flagger.AddFlags.
func (b *OptionAggregator) AddFlags(fs *pflag.FlagSet) {
	for _, option := range *b {
		option.AddFlags(fs)
	}
}

// Complete implements Completer.Complete.
func (b *OptionAggregator) Complete() error {
	for _, option := range *b {
		if err := option.Complete(); err != nil {
			return err
		}
	}
	return nil
}

// ManagerOptions are command line options that can be set for manager.Options.
type ManagerOptions struct {
	// LeaderElection is whether leader election is turned on or not.
	LeaderElection bool
	// LeaderElectionID is the id to do leader election with.
	LeaderElectionID string
	// LeaderElectionNamespace is the namespace to do leader election in.
	LeaderElectionNamespace string

	config *ManagerConfig
}

// AddFlags implements Flagger.AddFlags.
func (m *ManagerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&m.LeaderElection, LeaderElectionFlag, m.LeaderElection, "Whether to use leader election or not when running this controller manager.")
	fs.StringVar(&m.LeaderElectionID, LeaderElectionIDFlag, m.LeaderElectionID, "The leader election id to use.")
	fs.StringVar(&m.LeaderElectionNamespace, LeaderElectionNamespaceFlag, m.LeaderElectionNamespace, "The namespace to do leader election in.")
}

// Complete implements Completer.Complete.
func (m *ManagerOptions) Complete() error {
	m.config = &ManagerConfig{m.LeaderElection, m.LeaderElectionID, m.LeaderElectionNamespace}
	return nil
}

// Completed returns the completed ManagerConfig. Only call this if `Complete` was successful.
func (m *ManagerOptions) Completed() *ManagerConfig {
	return m.config
}

// ManagerConfig is a completed manager configuration.
type ManagerConfig struct {
	// LeaderElection is whether leader election is turned on or not.
	LeaderElection bool
	// LeaderElectionID is the id to do leader election with.
	LeaderElectionID string
	// LeaderElectionNamespace is the namespace to do leader election in.
	LeaderElectionNamespace string
}

// Apply sets the values of this ManagerConfig in the given manager.Options.
func (c *ManagerConfig) Apply(opts *manager.Options) {
	opts.LeaderElection = c.LeaderElection
	opts.LeaderElectionID = c.LeaderElectionID
	opts.LeaderElectionNamespace = c.LeaderElectionNamespace
}

// Options initializes empty manager.Options, applies the set values and returns it.
func (c *ManagerConfig) Options() manager.Options {
	var opts manager.Options
	c.Apply(&opts)
	return opts
}

// ControllerOptions are command line options that can be set for controller.Options.
type ControllerOptions struct {
	// MaxConcurrentReconciles are the maximum concurrent reconciles.
	MaxConcurrentReconciles int

	config *ControllerConfig
}

// AddFlags implements Flagger.AddFlags.
func (c *ControllerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&c.MaxConcurrentReconciles, MaxConcurrentReconcilesFlag, c.MaxConcurrentReconciles, "The maximum number of concurrent reconciliations.")
}

// Complete implements Completer.Complete.
func (c *ControllerOptions) Complete() error {
	c.config = &ControllerConfig{c.MaxConcurrentReconciles}
	return nil
}

// Completed returns the completed ControllerConfig. Only call this if `Complete` was successful.
func (c *ControllerOptions) Completed() *ControllerConfig {
	return c.config
}

// ControllerConfig is a completed controller configuration.
type ControllerConfig struct {
	// MaxConcurrentReconciles is the maximum number of concurrent reconciles.
	MaxConcurrentReconciles int
}

// Apply sets the values of this ControllerConfig in the given controller.Options.
func (c *ControllerConfig) Apply(opts *controller.Options) {
	opts.MaxConcurrentReconciles = c.MaxConcurrentReconciles
}

// Options initializes empty controller.Options, applies the set values and returns it.
func (c *ControllerConfig) Options() controller.Options {
	var opts controller.Options
	c.Apply(&opts)
	return opts
}

// RESTOptions are command line options that can be set for rest.Config.
type RESTOptions struct {
	// Kubeconfig is the path to a kubeconfig.
	Kubeconfig string
	// MasterURL is an override for the URL in a kubeconfig. Only used if out-of-cluster.
	MasterURL string

	config *RESTConfig
}

// RESTConfig is a completed REST configuration.
type RESTConfig struct {
	// Config is the rest.Config.
	Config *rest.Config
}

var (
	// BuildConfigFromFlags creates a build configuration from the given flags. Exposed for testing.
	BuildConfigFromFlags = clientcmd.BuildConfigFromFlags
	// InClusterConfig obtains the current in-cluster config. Exposed for testing.
	InClusterConfig = rest.InClusterConfig
	// Getenv obtains the environment variable with the given name. Exposed for testing.
	Getenv = os.Getenv
	// RecommendedHomeFile is the recommended location of the kubeconfig. Exposed for testing.
	RecommendedHomeFile = clientcmd.RecommendedHomeFile
)

func (r *RESTOptions) buildConfig() (*rest.Config, error) {
	// If a flag is specified with the config location, use that
	if len(r.Kubeconfig) > 0 {
		return BuildConfigFromFlags(r.MasterURL, r.Kubeconfig)
	}
	// If an env variable is specified with the config location, use that
	if kubeconfig := Getenv(clientcmd.RecommendedConfigPathEnvVar); len(kubeconfig) > 0 {
		return BuildConfigFromFlags(r.MasterURL, kubeconfig)
	}
	// If no explicit location, try the in-cluster config
	if c, err := InClusterConfig(); err == nil {
		return c, nil
	}

	return BuildConfigFromFlags("", RecommendedHomeFile)
}

// Complete implements RESTCompleter.Complete.
func (r *RESTOptions) Complete() error {
	config, err := r.buildConfig()
	if err != nil {
		return err
	}

	r.config = &RESTConfig{config}
	return nil
}

// Completed returns the completed RESTConfig. Only call this if `Complete` was successful.
func (r *RESTOptions) Completed() *RESTConfig {
	return r.config
}

// AddFlags implements Flagger.AddFlags.
func (r *RESTOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&r.Kubeconfig, KubeconfigFlag, "",
		"Paths to a kubeconfig. Only required if out-of-cluster.")
	fs.StringVar(&r.MasterURL, MasterURLFlag, "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. "+
			"Only required if out-of-cluster.")
}
