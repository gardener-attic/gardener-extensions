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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CloudProfileConfig contains provider-specific configuration that is embedded into Gardener's `CloudProfile`
// resource.
type CloudProfileConfig struct {
	metav1.TypeMeta `json:",inline"`
	// NamePrefix is used for naming NSX-T resources
	NamePrefix string `json:"namePrefix"`
	// Folder is the vSphere folder name to store the cloned machine VM (worker nodes)
	Folder string `json:"folder"`
	// Regions is the specification of regions and zones topology
	Regions []RegionSpec `json:"regions"`
	// DefaultClassStoragePolicyName is the name of the vSphere storage policy to use for the 'default-class' storage class
	DefaultClassStoragePolicyName string `json:"defaultClassStoragePolicyName"`
	// FailureDomainLabels are the tag categories used for regions and zones.
	// +optional
	FailureDomainLabels *FailureDomainLabels `json:"failureDomainLabels,omitempty"`
	// DNSServers is a list of IPs of DNS servers used while creating subnets.
	DNSServers []string `json:"dnsServers"`
	// MachineImages is the list of machine images that are understood by the controller. It maps
	// logical names and versions to provider-specific identifiers.
	MachineImages []MachineImages `json:"machineImages"`
	// Constraints is an object containing constraints for certain values in the control plane config.
	Constraints Constraints `json:"constraints"`
}

// FailureDomainLabels are the tag categories used for regions and zones in vSphere CSI driver and cloud controller.
// See Cloud Native Storage: Set Up Zones in the vSphere CNS Environment
// (https://docs.vmware.com/en/VMware-vSphere/6.7/Cloud-Native-Storage/GUID-9BD8CD12-CB24-4DF4-B4F0-A862D0C82C3B.html)
type FailureDomainLabels struct {
	// Region is the tag category used for region on vSphere data centers and/or clusters.
	Region string `json:"region"`
	// Zone is the tag category used for zones on vSphere data centers and/or clusters.
	Zone string `json:"zone"`
}

// RegionSpec specifies the topology of a region and its zones.
// A region consists of a Vcenter host, transport zone and optionally a data center.
// A zone in a region consists of a data center (if not specified in the region), a computer cluster,
// and optionally a resource zone or host system.
type RegionSpec struct {
	// Name is the name of the region
	Name string `json:"name"`
	// VsphereHost is the vSphere host
	VsphereHost string `json:"vsphereHost"`
	// VsphereInsecureSSL is a flag if insecure HTTPS is allowed for VsphereHost
	VsphereInsecureSSL bool `json:"vsphereInsecureSSL"`
	// NSXTHost is the NSX-T host
	NSXTHost string `json:"nsxtHost"`
	// NSXTInsecureSSL is a flag if insecure HTTPS is allowed for NSXTHost
	NSXTInsecureSSL bool `json:"nsxtInsecureSSL"`
	// TransportZone is the NSX-T transport zone
	TransportZone string `json:"transportZone"`
	// LogicalTier0Router is the NSX-T logical tier 0 router
	LogicalTier0Router string `json:"logicalTier0Router"`
	// EdgeCluster is the NSX-T edge cluster
	EdgeCluster string `json:"edgeCluster"`
	// SNATIPPool is the NSX-T IP pool to allocate the SNAT ip address
	SNATIPPool string `json:"snatIPPool"`

	// Datacenter is the name of the vSphere data center (data center can either be defined at region or zone level)
	// +optional
	Datacenter *string `json:"datacenter,omitempty"`

	// Datastore is the vSphere datastore to store the cloned machine VM. Either Datastore or DatastoreCluster must be specified at region or zones level.
	// +optional
	Datastore *string `json:"datastore,omitempty"`
	// DatastoreCluster is the vSphere  datastore cluster to store the cloned machine VM. Either Datastore or DatastoreCluster must be specified at region or zones level.
	// +optional
	DatastoreCluster *string `json:"datastoreCluster,omitempty"`

	// Zones is the list of zone specifications of the region.
	Zones []ZoneSpec `json:"zones"`

	// CaFile is the optional CA file to be trusted when connecting to vCenter. If not set, the node's CA certificates will be used. Only relevant if InsecureFlag=0
	// +optional
	CaFile *string `json:"caFile,omitempty"`
	// Thumbprint is the optional vCenter certificate thumbprint, this ensures the correct certificate is used
	// +optional
	Thumbprint *string `json:"thumbprint,omitempty"`

	// DNSServers is a optional list of IPs of DNS servers used while creating subnets. If provided, it overwrites the global
	// DNSServers of the CloudProfileConfig
	// +optional
	DNSServers []string `json:"dnsServers,omitempty"`
	// MachineImages is the list of machine images that are understood by the controller. If provided, it overwrites the global
	// MachineImages of the CloudProfileConfig
	// +optional
	MachineImages []MachineImages `json:"machineImages,omitempty"`
}

// ZoneSpec specifies a zone of a region.
// A zone in a region consists of a data center (if not specified in the region), a computer cluster,
// and optionally a resource zone or host system.
type ZoneSpec struct {
	// Nmae is the name of the zone
	Name string `json:"name"`
	// Datacenter is the name of the vSphere data center (data center can either be defined at region or zone level)
	// +optional
	Datacenter *string `json:"datacenter,omitempty"`

	// ComputeCluster is the name of the vSphere compute cluster. Either ComputeCluster or ResourcePool or HostSystem must be specified
	// +optional
	ComputeCluster *string `json:"computeCluster,omitempty"`
	// ResourcePool is the name of the vSphere resource pool. Either ComputeCluster or ResourcePool or HostSystem must be specified
	// +optional
	ResourcePool *string `json:"resourcePool,omitempty"`
	// HostSystem is the name of the vSphere host system. Either ComputeCluster or ResourcePool or HostSystem must be specified
	// +optional
	HostSystem *string `json:"hostSystem,omitempty"`

	// Datastore is the vSphere datastore to store the cloned machine VM. Either Datastore or DatastoreCluster must be specified at region or zones level.
	// +optional
	Datastore *string `json:"datastore,omitempty"`
	// DatastoreCluster is the vSphere  datastore cluster to store the cloned machine VM. Either Datastore or DatastoreCluster must be specified at region or zones level.
	// +optional
	DatastoreCluster *string `json:"datastoreCluster,omitempty"`
}

// Constraints is an object containing constraints for the shoots.
type Constraints struct {
	// LoadBalancerConfig contains constraints regarding allowed values of the 'Lo' block in the control plane config.
	LoadBalancerConfig LoadBalancerConfig `json:"loadBalancerConfig"`
}

// MachineImages is a mapping from logical names and versions to provider-specific identifiers.
type MachineImages struct {
	// Name is the logical name of the machine image.
	Name string `json:"name"`
	// Versions contains versions and a provider-specific identifier.
	Versions []MachineImageVersion `json:"versions"`
}

// MachineImageVersion contains a version and a provider-specific identifier.
type MachineImageVersion struct {
	// Version is the version of the image.
	Version string `json:"version"`
	// Path is the path of the VM template.
	Path string `json:"path"`
	// GuestID is the optional guestId to overwrite the guestId of the VM template.
	GuestID string `json:"guestId,omitempty"`
}

// LoadBalancerConfig contains the constraints for usable load balancer classes
type LoadBalancerConfig struct {
	// Size is the NSX-T load balancer size ("SMALL", "MEDIUM", or "LARGE")
	Size string `json:"size"`
	// Classes are the defined load balancer classes
	Classes []LoadBalancerClass `json:"classes"`
}

// LoadBalancerClass defines a restricted network setting for generic LoadBalancer classes.
type LoadBalancerClass struct {
	// Name is the name of the LB class
	Name string `json:"name"`
	// IPPoolName is the name of the NSX-T IP pool.
	IPPoolName string `json:"ipPoolName"`
}
