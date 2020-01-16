# Using the vSphere provider extension with Gardener as operator

The [`core.gardener.cloud/v1alpha1.CloudProfile` resource](https://github.com/gardener/gardener/blob/master/example/30-cloudprofile.yaml) declares a `providerConfig` field that is meant to contain provider-specific configuration.

In this document we are describing how this configuration looks like for VMware vSphere and provide an example `CloudProfile` manifest with minimal configuration that you can use to allow creating vSphere shoot clusters.

## `CloudProfileConfig`

The cloud profile configuration contains information about the real machine image paths in the vSphere environment (image names).
You have to map every version that you specify in `.spec.machineImages[].versions` here such that the vSphere extension knows the image ID for every version you want to offer.

It also contains optional default values for DNS servers that shall be used for shoots.
In the `dnsServers[]` list you can specify IP addresses that are used as DNS configuration for created shoot subnets.

Also, you have to specify several name of NSX-T objects in the constraints.

An example `CloudProfileConfig` for the vSphere extension looks as follows:

```yaml
apiVersion: vsphere.provider.extensions.gardener.cloud/v1alpha1
kind: CloudProfileConfig
namePrefix: my_gardener
defaultClassStoragePolicyName: "vSAN Default Storage Policy"
folder: my-vsphere-vm-folder
regions:
- name: region1
  vsphereHost: my.vsphere.host
  vsphereInsecureSSL: true
  nsxtHost: my.vsphere.host
  nsxtInsecureSSL: true
  transportZone: "my-tz"
  logicalTier0Router: "my-tier0router"
  edgeCluster: "my-edgecluster"
  snatIpPool: "my-snat-ip-pool"
  datacenter: my-vsphere-dc
  zones:
  - name: zone1
    computeCluster: my-vsphere-computecluster1
    # resourcePool: my-resource-pool1 # provide either computeCluster or resourcePool or hostSystem
    # hostSystem: my-host1 # provide either computeCluster or resourcePool or hostSystem
    datastore: my-vsphere-datastore1
    #datastoreCluster: my-vsphere-datastore-cluster # provide either datastore or datastoreCluster
  - name: zone2
    computeCluster: my-vsphere-computecluster2
    # resourcePool: my-resource-pool2 # provide either computeCluster or resourcePool or hostSystem
    # hostSystem: my-host2 # provide either computeCluster or resourcePool or hostSystem
    datastore: my-vsphere-datastore2
    #datastoreCluster: my-vsphere-datastore-cluster # provide either datastore or datastoreCluster
constraints:
  loadBalancerConfig:
    size: MEDIUM
    classes:
    - name: default
      ipPoolName: gardener_lb_vip
dnsServers:
- 10.10.10.11
- 10.10.10.12
machineImages:
- name: coreos
  versions:
  - version: 2191.5.0
    path: gardener/templates/coreos-2191.5.0
    guestId: coreos64Guest
```

## Example `CloudProfile` manifest

Please find below an example `CloudProfile` manifest:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: CloudProfile
metadata:
  name: vsphere
spec:
  type: vsphere
  providerConfig:
    apiVersion: vsphere.provider.extensions.gardener.cloud/v1alpha1
    kind: CloudProfileConfig
    namePrefix: my_gardener
    defaultClassStoragePolicyName: "vSAN Default Storage Policy"
    folder: my-vsphere-vm-folder
    regions:
    - name: region1
      vsphereHost: my.vsphere.host
      vsphereInsecureSSL: true
      nsxtHost: my.vsphere.host
      nsxtInsecureSSL: true
      transportZone: "my-tz"
      logicalTier0Router: "my-tier0router"
      edgeCluster: "my-edgecluster"
      snatIpPool: "my-snat-ip-pool"
      datacenter: my-vsphere-dc
      zones:
      - name: zone1
        computeCluster: my-vsphere-computecluster1
        # resourcePool: my-resource-pool1 # provide either computeCluster or resourcePool or hostSystem
        # hostSystem: my-host1 # provide either computeCluster or resourcePool or hostSystem
        datastore: my-vsphere-datastore1
        #datastoreCluster: my-vsphere-datastore-cluster # provide either datastore or datastoreCluster
      - name: zone2
        computeCluster: my-vsphere-computecluster2
        # resourcePool: my-resource-pool2 # provide either computeCluster or resourcePool or hostSystem
        # hostSystem: my-host2 # provide either computeCluster or resourcePool or hostSystem
        datastore: my-vsphere-datastore2
        #datastoreCluster: my-vsphere-datastore-cluster # provide either datastore or datastoreCluster
    constraints:
      loadBalancerConfig:
        size: MEDIUM
        classes:
        - name: default
          ipPoolName: gardener_lb_vip
    dnsServers:
    - 10.10.10.11
    - 10.10.10.12
    machineImages:
    - name: coreos
      versions:
      - version: 2191.5.0
        path: gardener/templates/coreos-2191.5.0
        guestId: coreos64Guest
  kubernetes:
    versions:
    - version: 1.15.4
    - version: 1.16.0
    - version: 1.16.1
  machineImages:
  - name: coreos
    versions:
    - version: 2191.5.0
  machineTypes:
  - name: std-02
    cpu: "2"
    gpu: "0"
    memory: 8Gi
    usable: true
  - name: std-04
    cpu: "4"
    gpu: "0"
    memory: 16Gi
    usable: true
  - name: std-08
    cpu: "8"
    gpu: "0"
    memory: 32Gi
    usable: true
  regions:
  - name: region1
    zones:
    - name: zone1
    - name: zone2
```

## Which versions of Kubernetes/vSphere are supported

This extension targets Kubernetes >= `v1.15` and vSphere `6.7 U3` or later.

- vSphere CSI driver needs vSphere `6.7 U3` or later,
  and Kubernetes >= `v1.14`
  (see [cloud-provider-vsphere CSI - Container Storage Interface](https://github.com/kubernetes/cloud-provider-vsphere/blob/master/docs/book/container_storage_interface.md#which-versions-of-kubernetesvsphere-support-it) )
- vSpere CPI driver needs vSphere `6.7 U3` or later,
  and Kubernetes >= `v1.11`
  (see [cloud-provider-vsphere CPI - Cloud Provider Interface](https://github.com/kubernetes/cloud-provider-vsphere/blob/master/docs/book/cloud_provider_interface.md#which-versions-of-kubernetesvsphere-support-it) )

## Supported VM images

Currently, only CoreOS and Flatcar (CoreOS fork) are supported.
Virtual Machine Hardware must be version 15 or higher, but images are upgraded
automatically if their hardware has an older version.
