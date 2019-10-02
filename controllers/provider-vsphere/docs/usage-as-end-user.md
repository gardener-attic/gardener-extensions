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

## `InfrastructureConfig`

The infrastructure configuration mainly describes how the network layout looks like in order to create the shoot worker nodes in a latter step, thus, prepares everything relevant to create VMs, load balancers, volumes, etc.

An example `InfrastructureConfig` for the vSphere extension looks as follows:

```yaml
infrastructureConfig:
  apiVersion: vsphere.provider.extensions.gardener.cloud/v1alpha1
  kind: InfrastructureConfig
  networks:
    worker: 10.250.0.0/19
```

The `floatingPoolName` is the name of the floating pool you want to use for your shoot.
If you don't know which floating pools are available look it up in the respective `CloudProfile`.

The `networks.router` section describes whether you want to create the shoot cluster in an already existing router or whether to create a new one:

* If `networks.router.name` is given then you have to specify the router name of the existing router that was created by other means (manually, other tooling, ...).
If you want to get a fresh router for the shoot then just omit the `networks.router` field.

The `networks.workers` section describes the CIDR for a subnet that is used for all shoot worker nodes, i.e., VMs which later run your applications.

You can freely choose these CIDRs and it is your responsibility to properly design the network layout to suit your needs.

Apart from the router and the worker subnet the vSphere extension will also create a network, router interfaces, security groups, and a key pair.

## `ControlPlaneConfig`

The control plane configuration mainly contains values for the vSphere-specific control plane components.
Today, the only component deployed by the vSphere extension is the `cloud-controller-manager`.

An example `ControlPlaneConfig` for the vSphere extension looks as follows:

```yaml
apiVersion: vsphere.provider.extensions.gardener.cloud/v1alpha1
kind: ControlPlaneConfig
cloudControllerManager:
  featureGates:
    CustomResourceValidation: true
```

The `cloudControllerManager.featureGates` contains a map of explicitly enabled or disabled feature gates.
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
    infrastructureConfig:
      apiVersion: vsphere.provider.extensions.gardener.cloud/v1alpha1
      kind: InfrastructureConfig
      networks:
        worker: 10.250.0.0/19
    controlPlaneConfig:
      apiVersion: vsphere.provider.extensions.gardener.cloud/v1alpha1
      kind: ControlPlaneConfig
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
