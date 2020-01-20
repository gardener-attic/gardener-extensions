# Using the vSphere provider extension with Gardener as end-user

The [`core.gardener.cloud/v1alpha1.Shoot` resource](https://github.com/gardener/gardener/blob/master/example/90-shoot.yaml) declares a few fields that are meant to contain provider-specific configuration.

In this document we are describing how this configuration looks like for VMware vSphere and provide an example `Shoot` manifest with minimal configuration that you can use to create an vSphere cluster (modulo the landscape-specific information like cloud profile names, secret binding names, etc.).

## Provider secret data

Every shoot cluster references a `SecretBinding` which itself references a `Secret`, and this `Secret` contains the provider credentials of your vSphere tenant.
It contains two authentication sets. One for the vSphere host and another for the NSX-T host, which is needed to set up the network infrastructure.
This `Secret` must look as follows:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: core-vsphere
  namespace: garden-dev
type: Opaque
data:
  vsphereHost: base64(vsphere-host)
  vsphereUsername: base64(vsphere-username)
  vspherePassword: base64(vsphere-password)
  vsphereInsecureSSL: base64("true"|"false")
  nsxtHost: base64(NSX-T-host)
  nsxtUsername: base64(NSX-T-username)
  nsxtPassword: base64(NSX-T-password)
  nsxtInsecureSSL: base64("true"|"false")
```

Here `base64(...)` are only a placeholders for the Base64 encoded values.

## `InfrastructureConfig`

The infrastructure configuration is currently not used. Nodes on all zones are using IP addresses from the common nodes
network as the network is managed by NSX-T.

An example `InfrastructureConfig` for the vSphere extension looks as follows (currently always empty):

```yaml
infrastructureConfig:
  apiVersion: vsphere.provider.extensions.gardener.cloud/v1alpha1
  kind: InfrastructureConfig
```

The infrastructure controller will create several network objects using NSX-T. A logical switch to be used as the network
for the VMs (nodes), a tier-1 router, a DHCP server, and a SNAT for the nodes. 

## `ControlPlaneConfig`

The control plane configuration mainly contains values for the vSphere-specific control plane components.
Today, the only component deployed by the vSphere extension is the `cloud-controller-manager`.

An example `ControlPlaneConfig` for the vSphere extension looks as follows:

```yaml
apiVersion: vsphere.provider.extensions.gardener.cloud/v1alpha1
kind: ControlPlaneConfig
loadBalancerClasses:
  - name: mypubliclbclass
  - name: myprivatelbclass
    ipPoolName: pool42 # optional overwrite
cloudControllerManager:
  featureGates:
    CustomResourceValidation: true
```

The `loadBalancerClasses` optionally defines the load balancer classes to be used.
The specified names must be defined in the constraints section of the cloud profile.
If the list contains a load balancer named "default", it is used as the default load balancer.
Otherwise the first one is also the default.
If no classes are specified the default load balancer class is used as defined in the cloud profile constraints section.

The `cloudControllerManager.featureGates` contains an optional map of explicitly enabled or disabled feature gates.
For production usage it's not recommend to use this field at all as you can enable alpha features or disable beta/stable features, potentially impacting the cluster stability.
If you don't want to configure anything for the `cloudControllerManager` simply omit the key in the YAML specification.

## Example `Shoot` manifest (one availability zone)

Please find below an example `Shoot` manifest for one availability zone:

```yaml
apiVersion: core.gardener.cloud/v1alpha1
kind: Shoot
metadata:
  name: johndoe-vsphere
  namespace: garden-dev
spec:
  cloudProfileName: vsphere
  region: europe-1
  secretBindingName: core-vsphere
  provider:
    type: vsphere
   
    ## infrastructureConfig is currently unused
    #infrastructureConfig:
    #  apiVersion: vsphere.provider.extensions.gardener.cloud/v1alpha1
    #  kind: InfrastructureConfig

    ## controlPlaneConfig has only optional parameters. Uncomment the following lines if needed
    #controlPlaneConfig:
    #  apiVersion: vsphere.provider.extensions.gardener.cloud/v1alpha1
    #  kind: ControlPlaneConfig
    #  loadBalancerClasses:
    #  - name: mylbclass

    workers:
    - name: worker-xoluy
      machine:
        type: std-04
      minimum: 2
      maximum: 2
      zones:
      - europe-1a
  networking:
    nodes: 10.250.0.0/16
    type: calico
  kubernetes:
    version: 1.16.1
  maintenance:
    autoUpdate:
      kubernetesVersion: true
      machineImageVersion: true
  addons:
    kubernetes-dashboard:
      enabled: true
    nginx-ingress:
      enabled: true
```
