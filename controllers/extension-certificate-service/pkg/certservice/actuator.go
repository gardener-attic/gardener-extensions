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

package certservice

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/certservice/internal"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apiserver/pkg/authentication/user"

	"github.com/gardener/gardener/pkg/utils/secrets"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/imagevector"

	"github.com/gardener/gardener/pkg/chartrenderer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/gardener/gardener-extensions/pkg/controller"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/utils"
	"github.com/gardener/gardener-extensions/pkg/util"
	"github.com/gardener/gardener/pkg/client/kubernetes"

	"github.com/go-logr/logr"

	"github.com/gardener/gardener-extensions/pkg/controller/extension"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ActuatorName is the name of the Certificate Service actuator.
const ActuatorName = "certificate-service-actuator"

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator() extension.Actuator {
	return &actuator{
		logger: log.Log.WithName(ActuatorName),
	}
}

type actuator struct {
	client client.Client
	config *rest.Config

	logger logr.Logger
}

// RBACManagerSecretName is the name of a secret which contains the Kubeconfig for the RBAC manager.
const RBACManagerSecretName = "cert-service-rbac-manager"

// CreateOrUpdate the Extension resource.
func (a *actuator) CreateOrUpdate(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()
	cluster, err := controller.GetCluster(ctx, a.client, namespace)
	if err != nil {
		return err
	}

	if err := a.deployCertBroker(ctx, cluster.Shoot, namespace); err != nil {
		return err
	}

	certConfig := secrets.CertificateSecretConfig{
		Name:         RBACManagerSecretName,
		CommonName:   "system:" + RBACManagerSecretName,
		Organization: []string{user.SystemPrivilegedGroup},
	}

	if _, err := util.GetOrCreateShootKubeconfig(ctx, a.client, certConfig, namespace); err != nil {
		return fmt.Errorf("error setting up rbac manager: %v", err)
	}

	return nil
}

// Delete the Extension resource.
func (a *actuator) Delete(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()
	if err := a.deleteCertBroker(ctx, namespace); err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RBACManagerSecretName,
			Namespace: namespace,
		},
	}

	if err := a.client.Delete(ctx, secret); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("error cleaning up secret for rbac manager: %v", err)
	}

	return nil
}

// InjectConfig injects the rest config to this actuator.
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config
	return nil
}

// InjectClient injects the controller runtime client into the reconciler.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

const (
	certBrokerCN   = "garden.sapcloud.io:system:cert-broker"
	certBrokerName = "cert-broker"
)

var certBrokerChartPath = filepath.Join(utils.ChartsPath, certBrokerName)

func (a *actuator) deployCertBroker(ctx context.Context, shoot *gardenv1beta1.Shoot, namespace string) error {
	config, err := utils.RetrieveConfig(ctx, a.client, ResourceName, ResourceNamespace)
	if err != nil {
		return fmt.Errorf("error retrieving certificate service config: %v", err)
	}

	shootDomain := shoot.Spec.DNS.Domain
	if shootDomain == nil {
		return fmt.Errorf("no domain given for shoot %s/%s", shoot.GetName(), shoot.GetNamespace())
	}

	var (
		dns        []map[string]string
		configSpec = config.Spec
	)

	for _, route53Provider := range configSpec.Providers.Route53 {
		route53values := internal.CreateDNSProviderValue(&route53Provider, *shootDomain)
		if route53values != nil {
			dns = append(dns, route53values)
		}
	}
	for _, cloudDNSProvider := range configSpec.Providers.CloudDNS {
		cloudDNSValues := internal.CreateDNSProviderValue(&cloudDNSProvider, *shootDomain)
		dns = append(dns, cloudDNSValues)
	}

	certConfig := secrets.CertificateSecretConfig{
		Name:       certBrokerName,
		CommonName: certBrokerCN,
	}

	shootKubeconfig, err := util.GetOrCreateShootKubeconfig(ctx, a.client, certConfig, namespace)
	if err != nil {
		return err
	}

	shootKubeconfigChecksum := util.ComputeSecretCheckSum(shootKubeconfig.Data)

	certBrokerConfig := map[string]interface{}{
		"replicas": util.GetReplicaCount(shoot, 1),
		"certbroker": map[string]interface{}{
			"targetClusterSecret": shootKubeconfig.GetName(),
		},
		"certmanager": map[string]interface{}{
			"clusterissuer": configSpec.IssuerName,
			"dns":           dns,
		},
		"podAnnotations": map[string]interface{}{
			"checksum/secret-cert-broker": shootKubeconfigChecksum,
		},
	}

	certBrokerConfig, err = imagevector.ImageVector.InjectImages(certBrokerConfig, "", "", certBrokerName)
	if err != nil {
		return fmt.Errorf("failed to find image version for %s: %v", certBrokerName, err)
	}

	applier, err := kubernetes.NewChartApplierForConfig(a.config)
	if err != nil {
		return fmt.Errorf("failed to create chart applier: %v", err)
	}

	a.logger.Info("Component is being applied", "component", "cert-broker", "namespace", namespace)
	return applier.ApplyChartInNamespace(
		ctx,
		certBrokerChartPath,
		namespace,
		certBrokerName,
		certBrokerConfig,
		nil,
	)
}

func (a *actuator) deleteCertBroker(ctx context.Context, namespace string) error {
	chartRenderer, err := chartrenderer.NewForConfig(a.config)
	if err != nil {
		return err
	}
	configValues := make(map[string]interface{})
	release, err := chartRenderer.Render(
		certBrokerChartPath,
		certBrokerName,
		namespace,
		configValues)
	if err != nil {
		return fmt.Errorf("failed to render chart: %v", err)
	}

	var (
		manifest   = release.Manifest()
		decoder    = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 1024)
		decodedObj map[string]interface{}
	)

	a.logger.Info("Component is being deleted", "component", "cert-broker", "namespace", namespace)
	for err = decoder.Decode(&decodedObj); err == nil; err = decoder.Decode(&decodedObj) {
		if decodedObj == nil {
			continue
		}

		obj := unstructured.Unstructured{Object: decodedObj}
		decodedObj = nil

		// only delete namepaced resources
		if len(obj.GetNamespace()) > 0 {
			if err := a.client.Delete(ctx, &obj); err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}

	}
	return nil
}
