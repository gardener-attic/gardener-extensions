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

package openstack

import "path/filepath"

const (
	// Name is the name of the OpenStack provider.
	Name = "provider-openstack"
	// StorageProviderName is the name of the Openstack storage provider.
	StorageProviderName = "Swift"

	// MachineControllerManagerImageName is the name of the MachineControllerManager image.
	MachineControllerManagerImageName = "machine-controller-manager"
	// CloudControllerImageName is the name of the external OpenStackCloudProvider image.
	CloudControllerImageName = "openstack-cloud-controller-manager"
	// ETCDBackupRestoreImageName is the name of the etcd backup and restore image.
	ETCDBackupRestoreImageName = "etcd-backup-restore"

	// AuthURL is a constant for the key in a cloud provider secret that holds the OpenStack auth url.
	AuthURL = "authURL"
	// DomainName is a constant for the key in a cloud provider secret that holds the OpenStack domain name.
	DomainName = "domainName"
	// DomainID is a constant for the key in a cloud provider secret that holds the OpenStack domain name.
	DomainID = "domainID"
	// TenantName is a constant for the key in a cloud provider secret that holds the OpenStack tenant name.
	TenantName = "tenantName"
	// TenantID is a constant for the key in a cloud provider secret that holds the OpenStack tenant name.
	TenantID = "tenantID"
	// UserDomainName is a constant for the key in a cloud provider secret that holds the OpenStack tenant name.
	UserDomainName = "userDomainName"
	// UserDomainID is a constant for the key in a cloud provider secret that holds the OpenStack tenant name.
	UserDomainID = "userDomainID"
	// UserName is a constant for the key in a cloud provider secret and backup secret that holds the OpenStack username.
	UserName = "username"
	// Password is a constant for the key in a cloud provider secret and backup secret that holds the OpenStack password.
	Password = "password"
	// Region is a constant for the key in a backup secret that holds the Openstack region.
	Region = "region"
	// CloudYAML is a constant for the key in the backup or cloud provider secret that holds the access information for
	// the etcd backup infrastructure.
	CloudYAML = "clouds.yaml"
	// CloudYAMLKey is a constant for the credential key inside the CloudYAML.
	CloudYAMLKey = "cloudYAMLKey"
	// CloudYAMLDefaultKey is a constant for the default credentials key inside the CloudYAML.
	CloudYAMLDefaultKey = "openstack"
	// BucketName is a constant for the key in a backup secret that holds the bucket name.
	// The bucket name is written to the backup secret by Gardener as a temporary solution.
	// TODO In the future, the bucket name should come from a BackupBucket resource (see https://github.com/gardener/gardener/blob/master/docs/proposals/02-backupinfra.md)
	BucketName = "bucketName"

	// CloudProviderConfigCloudControllerManagerName is the name of the configmap containing the cloud provider config.
	CloudProviderConfigCloudControllerManagerName = "cloud-provider-config-cloud-controller-manager"
	// CloudProviderConfigKubeControllerManagerName is the name of the original configmap containing the cloud provider config (for compatibility reasons).
	CloudProviderConfigKubeControllerManagerName = "cloud-provider-config-kube-controller-manager"
	// CloudProviderConfigMapKey is the key storing the cloud provider config as value in the cloud provider configmap.
	CloudProviderConfigMapKey = "cloudprovider.conf"
	// MachineControllerManagerName is a constant for the name of the machine-controller-manager.
	MachineControllerManagerName = "machine-controller-manager"
	// MachineControllerManagerVpaName is the name of the VerticalPodAutoscaler of the machine-controller-manager deployment.
	MachineControllerManagerVpaName = "machine-controller-manager-vpa"
	// BackupSecretName defines the name of the secret containing the credentials which are required to
	// authenticate against the respective cloud provider (required to store the backups of Shoot clusters).
	BackupSecretName = "etcd-backup"
)

var (
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("controllers", Name, "charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")
)
