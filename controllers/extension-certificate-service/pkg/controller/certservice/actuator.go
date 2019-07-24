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
	"context"
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/apis/config"
	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/controller/certservice/internal"
	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/imagevector"
	"github.com/gardener/gardener-extensions/controllers/extension-certificate-service/pkg/utils"
	"github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/extension"
	"github.com/gardener/gardener-extensions/pkg/util"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// ActuatorName is the name of the Certificate Service actuator.
const ActuatorName = "certificate-service-actuator"

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator(config config.Configuration) extension.Actuator {
	return &actuator{
		logger:            log.Log.WithName(ActuatorName),
		certServiceConfig: config,
	}
}

type actuator struct {
	applier kubernetes.ChartApplier
	client  client.Client
	config  *rest.Config

	certServiceConfig config.Configuration

	logger logr.Logger
}

// Reconcile the Extension resource.
func (a *actuator) Reconcile(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()

	secret, err := util.GetGardenerSecret(ctx, a.client, namespace)
	if err != nil {
		return fmt.Errorf("Error getting Gardener secret from namespace %s: %v", namespace, err)
	}

	kubecfg, err := util.GetKubeconfigFromSecret(secret)
	if err != nil {
		return fmt.Errorf("Error creating Kubeconfig to deploy RBAC manifests: %v", err)
	}

	cluster, err := controller.GetCluster(ctx, a.client, namespace)
	if err != nil {
		return err
	}

	dns := cluster.Shoot.Spec.DNS
	if dns.Domain == nil && dns.Provider != nil && *dns.Provider == gardenv1beta1.DNSUnmanaged {
		return nil
	}

	if !controller.IsHibernated(cluster.Shoot) {
		if err := a.createRBAC(ctx, kubecfg); err != nil {
			return err
		}
	}

	return a.createCertBroker(ctx, cluster.Shoot, namespace)
}

// Delete the Extension resource.
func (a *actuator) Delete(ctx context.Context, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()
	if err := a.deleteCertBroker(ctx, namespace); err != nil {
		return err
	}

	secret, err := util.GetGardenerSecret(ctx, a.client, namespace)
	if err != nil {
		return fmt.Errorf("Error getting Gardener secret from namespace %s: %v", namespace, err)
	}

	kubecfg, err := clientcmd.RESTConfigFromKubeConfig(secret.Data["kubeconfig"])
	if err != nil {
		return fmt.Errorf("Error creating Kubeconfig to delete RBAC manifests: %v", err)
	}

	if err := a.deleteRBAC(ctx, kubecfg); err != nil {
		return err
	}

	return nil
}

// InjectConfig injects the rest config to this actuator.
func (a *actuator) InjectConfig(config *rest.Config) error {
	a.config = config
	applier, err := kubernetes.NewChartApplierForConfig(a.config)
	if err != nil {
		return fmt.Errorf("failed to create chart applier: %v", err)
	}
	a.applier = applier
	return nil
}

// InjectClient injects the controller runtime client into the reconciler.
func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

func (a *actuator) createCertBroker(ctx context.Context, shoot *gardenv1beta1.Shoot, namespace string) error {
	shootDomain := shoot.Spec.DNS.Domain
	if shootDomain == nil {
		return fmt.Errorf("no domain given for shoot %s/%s", shoot.GetName(), shoot.GetNamespace())
	}

	var (
		dns        []map[string]string
		configSpec = a.certServiceConfig.Spec
	)

	for _, route53Provider := range configSpec.Providers.Route53 {
		if route53values := internal.CreateDNSProviderValue(&route53Provider, *shootDomain); route53values != nil {
			dns = append(dns, route53values)
		}
	}
	for _, cloudDNSProvider := range configSpec.Providers.CloudDNS {
		if cloudDNSValues := internal.CreateDNSProviderValue(&cloudDNSProvider, *shootDomain); cloudDNSValues != nil {
			dns = append(dns, cloudDNSValues)
		}
	}

	shootKubeconfig, err := a.createKubeconfigForCertManager(ctx, namespace)
	if err != nil {
		return err
	}

	shootKubeconfigChecksum := util.ComputeChecksum(shootKubeconfig.Data)

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

	certBrokerConfig, err = chart.InjectImages(certBrokerConfig, imagevector.ImageVector(), []string{utils.CertBrokerImageName})
	if err != nil {
		return fmt.Errorf("failed to find image version for %s: %v", utils.CertBrokerImageName, err)
	}

	a.logger.Info("Component is being applied", "component", "cert-broker", "namespace", namespace)
	return a.applier.ApplyChartInNamespace(
		ctx,
		filepath.Join(utils.ChartsPath, utils.CertBrokerResourceName),
		namespace,
		utils.CertBrokerResourceName,
		certBrokerConfig,
		nil,
	)
}

func (a *actuator) deleteCertBroker(ctx context.Context, namespace string) error {
	meta := metav1.ObjectMeta{
		Name:      utils.CertBrokerResourceName,
		Namespace: namespace,
	}

	objects := []runtime.Object{
		&appsv1.Deployment{
			ObjectMeta: meta,
		},
		&corev1.Secret{
			ObjectMeta: meta,
		},
		&corev1.ServiceAccount{
			ObjectMeta: meta,
		},
		&rbacv1.RoleBinding{
			ObjectMeta: meta,
		},
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: meta,
		},
	}

	a.logger.Info("Component is being deleted", "component", "cert-broker", "namespace", namespace)
	for _, obj := range objects {
		if err := a.client.Delete(ctx, obj); client.IgnoreNotFound(err) != nil {
			return err
		}
	}

	return nil
}

func (a *actuator) createRBAC(ctx context.Context, config *rest.Config) error {
	applier, err := kubernetes.NewChartApplierForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create chart applier: %v", err)
	}

	a.logger.Info("Manifests are being applied")
	if err := applier.ApplyChart(ctx, filepath.Join(utils.ChartsPath, "cert-broker-rbac"), "", "", nil, nil); err != nil {
		return fmt.Errorf("error applying manifests: %v", err)
	}

	return nil
}

func (a *actuator) deleteRBAC(ctx context.Context, config *rest.Config) error {
	meta := metav1.ObjectMeta{
		Name: utils.RBACResourceName,
	}

	objects := []runtime.Object{
		&rbacv1.ClusterRole{
			ObjectMeta: meta,
		},
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: meta,
		},
	}

	targetClient, err := client.New(config, client.Options{})
	if err != nil {
		return fmt.Errorf("Cannot create client for target cluster to delete RBAC manifests: %v", err)
	}

	for _, obj := range objects {
		if err := targetClient.Delete(ctx, obj); client.IgnoreNotFound(err) != nil {
			return err
		}
	}

	return nil
}

func (a *actuator) createKubeconfigForCertManager(ctx context.Context, namespace string) (*corev1.Secret, error) {
	certConfig := secrets.CertificateSecretConfig{
		Name:       utils.CertBrokerResourceName,
		CommonName: utils.CertBrokerUserName,
	}

	return util.GetOrCreateShootKubeconfig(ctx, a.client, certConfig, namespace)
}
