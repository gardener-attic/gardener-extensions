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

package coreos

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig"

	"github.com/gardener/gardener-extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"github.com/go-logr/logr"
	yaml "gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var coreOSCloudInitCommand = fmt.Sprintf("/usr/bin/coreos-cloudinit --from-file=")

type actuator struct {
	client client.Client
	scheme *runtime.Scheme
	logger logr.Logger
}

// NewActuator creates a new Actuator that updates the status of the handled OperatingSystemConfigs.
func NewActuator() operatingsystemconfig.Actuator {
	return &actuator{logger: log.Log.WithName("coreos-operatingsystemconfig-actuator")}
}

func (c *actuator) InjectScheme(scheme *runtime.Scheme) error {
	c.scheme = scheme
	return nil
}

func (c *actuator) InjectClient(client client.Client) error {
	c.client = client
	return nil
}

func (c *actuator) Exists(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) (bool, error) {
	return config.Status.CloudConfig != nil, nil
}

func (c *actuator) Create(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) error {
	return c.reconcile(ctx, config)
}

func (c *actuator) Update(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) error {
	return c.reconcile(ctx, config)
}

func (c *actuator) Delete(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) error {
	return c.delete(ctx, config)
}

func (c *actuator) reconcile(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) error {
	cloudConfig, units, err := c.cloudConfigFromOperatingSystemConfig(ctx, config)
	if err != nil {
		config.Status.ObservedGeneration = config.Generation
		config.Status.LastOperation, config.Status.LastError = controller.ReconcileError(extensionsv1alpha1.LastOperationTypeReconcile, fmt.Sprintf("Could not generate cloud config: %v", err), 50)
		if err := c.client.Status().Update(ctx, config); err != nil {
			c.logger.Error(err, "Could not update operating system config status after update error", "osc", config.Name)
		}
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: secretObjectMetaForConfig(config),
	}

	if err := controller.CreateOrUpdate(ctx, c.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[extensionsv1alpha1.OperatingSystemConfigSecretDataKey] = []byte(cloudConfig)

		return controllerutil.SetControllerReference(config, secret, c.scheme)
	}); err != nil {
		config.Status.ObservedGeneration = config.Generation
		config.Status.LastOperation, config.Status.LastError = controller.ReconcileError(extensionsv1alpha1.LastOperationTypeReconcile, fmt.Sprintf("Could not apply secret for generated cloud config: %v", err), 50)
		if err := c.client.Status().Update(ctx, config); err != nil {
			c.logger.Error(err, "Could not update operating system config status after reconcile error", "osc", config.Name)
		}
		return err
	}

	config.Status.CloudConfig = &extensionsv1alpha1.CloudConfig{
		SecretRef: corev1.SecretReference{
			Name:      secret.Name,
			Namespace: secret.Namespace,
		},
	}
	if path := config.Spec.ReloadConfigFilePath; path != nil {
		command := coreOSCloudInitCommand + *path
		config.Status.Command = &command
	}
	config.Status.Units = units
	config.Status.ObservedGeneration = config.Generation
	config.Status.LastOperation, config.Status.LastError = controller.ReconcileSucceeded(extensionsv1alpha1.LastOperationTypeReconcile, "Successfully generated cloud config")
	return c.client.Status().Update(ctx, config)
}

func (c *actuator) delete(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) error {
	config.Status.ObservedGeneration = config.Generation
	config.Status.LastOperation, config.Status.LastError = controller.ReconcileSucceeded(extensionsv1alpha1.LastOperationTypeDelete, "Successfully deleted cloud config")
	if err := c.client.Status().Update(ctx, config); err != nil {
		c.logger.Error(err, "Could not update operating system config status for deletion", "osc", config.Name)
		return err
	}
	return nil
}

func secretObjectMetaForConfig(config *extensionsv1alpha1.OperatingSystemConfig) metav1.ObjectMeta {
	var (
		name      = fmt.Sprintf("osc-result-%s", config.Name)
		namespace = config.Namespace
	)

	if cloudConfig := config.Status.CloudConfig; cloudConfig != nil {
		name = cloudConfig.SecretRef.Name
		namespace = cloudConfig.SecretRef.Namespace
	}

	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}
}

func (c *actuator) cloudConfigFromOperatingSystemConfig(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) (string, []string, error) {
	cloudConfig := &CloudConfig{
		CoreOS: Config{
			Update: Update{
				RebootStrategy: "off",
			},
			Units: []Unit{
				{
					Name: "update-engine.service",
					Mask: true,
				},
				{
					Name: "locksmithd.service",
					Mask: true,
				},
			},
		},
	}

	unitNames := make([]string, 0, len(config.Spec.Units))
	for _, unit := range config.Spec.Units {
		unitNames = append(unitNames, unit.Name)

		u := Unit{Name: unit.Name}

		if unit.Command != nil {
			u.Command = *unit.Command
		}
		if unit.Enable != nil {
			u.Enable = *unit.Enable
		}
		if unit.Content != nil {
			u.Content = *unit.Content
		}

		for _, dropIn := range unit.DropIns {
			u.DropIns = append(u.DropIns, UnitDropIn{
				Name:    dropIn.Name,
				Content: dropIn.Content,
			})
		}

		cloudConfig.CoreOS.Units = append(cloudConfig.CoreOS.Units, u)
	}

	for _, file := range config.Spec.Files {
		f := File{
			Path: file.Path,
		}

		permissions := extensionsv1alpha1.OperatingSystemConfigDefaultFilePermission
		if p := file.Permissions; p != nil {
			permissions = *p
		}
		f.RawFilePermissions = strconv.FormatInt(int64(permissions), 8)

		if file.Content.Inline != nil {
			f.Encoding = file.Content.Inline.Encoding
			f.Content = file.Content.Inline.Data
		}

		if file.Content.SecretRef != nil {
			var secret corev1.Secret
			if err := c.client.Get(ctx, client.ObjectKey{Name: file.Content.SecretRef.Name, Namespace: config.Namespace}, &secret); err != nil {
				return "", nil, err
			}

			data, ok := secret.Data[file.Content.SecretRef.DataKey]
			if !ok {
				return "", nil, fmt.Errorf("could not find key %q in data of secret %q", file.Content.SecretRef.DataKey, file.Content.SecretRef.Name)
			}

			f.Encoding = "b64"
			f.Content = base64.StdEncoding.EncodeToString(data)
		}

		cloudConfig.WriteFiles = append(cloudConfig.WriteFiles, f)
	}

	data, err := cloudConfig.String()
	if err != nil {
		return "", nil, err
	}

	return data, unitNames, nil
}

// String returns the string representation of the CloudConfig structure.
func (c CloudConfig) String() (string, error) {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("#cloud-config\n\n%s", string(bytes)), nil
}
