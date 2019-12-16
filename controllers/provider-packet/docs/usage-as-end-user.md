# Using the Packet provider extension with Gardener as end-user

The [`core.gardener.cloud/v1beta1.Shoot` resource](https://github.com/gardener/gardener/blob/master/example/90-shoot.yaml) declares a few fields that are meant to contain provider-specific configuration.

In this document we are describing how this configuration looks like for Packet and provide an example `Shoot` manifest with minimal configuration that you can use to create an Packet cluster (modulo the landscape-specific information like cloud profile names, secret binding names, etc.).

## Provider secret data

Every shoot cluster references a `SecretBinding` which itself references a `Secret`, and this `Secret` contains the provider credentials of your Packet project.
This `Secret` must look as follows:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: core-packet
  namespace: garden-dev
type: Opaque
data:
  apiToken: base64(api-token)
  projectID: base64(project-id)
```

Please look up https://www.packet.com/developers/api/ as well.

## `InfrastructureConfig`

Currently, there is no infrastructure configuration possible for the Packet environment.

An example `InfrastructureConfig` for the Packet extension looks as follows:

```yaml
apiVersion: packet.provider.extensions.gardener.cloud/v1alpha1
kind: InfrastructureConfig
```

The Packet extension will only create a key pair.

## `ControlPlaneConfig`

The control plane configuration mainly contains values for the Packet-specific control plane components.
Today, the Packet extension deploys the `cloud-controller-manager` and the CSI controllers, however, it doesn't offer any configuration options at the moment.

An example `ControlPlaneConfig` for the Packet extension looks as follows:

```yaml
apiVersion: packet.provider.extensions.gardener.cloud/v1alpha1
kind: ControlPlaneConfig
```

## Example `Shoot` manifest

Please find below an example `Shoot` manifest:

```yaml
apiVersion: core.gardener.cloud/v1alpha1
kind: Shoot
metadata:
  name: johndoe-packet
  namespace: garden-dev
spec:
  cloudProfileName: packet
  region: EWR1
  secretBindingName: core-packet
  provider:
    type: packet
    infrastructureConfig:
      apiVersion: packet.provider.extensions.gardener.cloud/v1alpha1
      kind: InfrastructureConfig
    controlPlaneConfig:
      apiVersion: packet.provider.extensions.gardener.cloud/v1alpha1
      kind: ControlPlaneConfig
    workers:
    - name: worker-xoluy
      machine:
        type: t1.small
      minimum: 2
      maximum: 2
      volume:
        size: 50Gi
        type: storage_1
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
