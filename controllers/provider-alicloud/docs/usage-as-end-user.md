# Using the Alicloud provider extension with Gardener as end-user

The [`core.gardener.cloud/v1alpha1.Shoot` resource](https://github.com/gardener/gardener/blob/master/example/90-shoot.yaml) declares a few fields that are meant to contain provider-specific configuration.

In this document we are describing how this configuration looks like for Alicloud and provide an example `Shoot` manifest with minimal configuration that you can use to create an Alicloud cluster (modulo the landscape-specific information like cloud profile names, secret binding names, etc.).

## Provider secret data

Every shoot cluster references a `SecretBinding` which itself references a `Secret`, and this `Secret` contains the provider credentials of your Alicloud account.
This `Secret` must look as follows:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: core-alicloud
  namespace: garden-dev
type: Opaque
data:
  accessKeyID: base64(access-key-id)
  accessKeySecret: base64(access-key-secret)
```

Please look up https://www.alibabacloud.com/help/doc-detail/29009.htm as well.

## `InfrastructureConfig`

The infrastructure configuration mainly describes how the network layout looks like in order to create the shoot worker nodes in a latter step, thus, prepares everything relevant to create VMs, load balancers, volumes, etc.

An example `InfrastructureConfig` for the Alicloud extension looks as follows:

```yaml
apiVersion: alicloud.provider.extensions.gardener.cloud/v1alpha1
kind: InfrastructureConfig
networks:
  vpc: # specify either 'id' or 'cidr'
  # id: my-vpc
    cidr: 10.250.0.0/16
  zones:
  - name: eu-central-1a
    worker: 10.250.1.0/24
```

The `networks.vpc` section describes whether you want to create the shoot cluster in an already existing VPC or whether to create a new one:

* If `networks.vpc.id` is given then you have to specify the VPC ID of the existing VPC that was created by other means (manually, other tooling, ...).
* If `networks.vpc.cidr` is given then you have to specify the VPC CIDR of a new VPC that will be created during shoot creation.
You can freely choose a private CIDR range.
* Either `networks.vpc.id` or `networks.vpc.cidr` must be present, but not both at the same time.

The `networks.zones` section describes which subnets you want to create in availability zones.
For every zone, the Alicloud extension creates one subnet:

* The `worker` subnet is used for all shoot worker nodes, i.e., VMs which later run your applications.

For every subnet, you have to specify a CIDR range contained in the VPC CIDR specified above, or the VPC CIDR of your already existing VPC.
You can freely choose these CIDR and it is your responsibility to properly design the network layout to suit your needs.

If you want to use multiple availability zones then add a second, third, ... entry to the `networks.zones[]` list and properly specify the AZ name in `networks.zones[].name`.

Apart from the VPC and the subnets the Alicloud extension will also create a NAT gateway (only if a new VPC is created), a key pair, elastic IPs, VSwitches, a SNAT table entry, and security groups.

## `ControlPlaneConfig`

The control plane configuration mainly contains values for the Alicloud-specific control plane components.
Today, the Alicloud extension deploys the `cloud-controller-manager` and the CSI controllers.

An example `ControlPlaneConfig` for the Alicloud extension looks as follows:

```yaml
apiVersion: alicloud.provider.extensions.gardener.cloud/v1alpha1
kind: ControlPlaneConfig
zone: eu-central-1a
cloudControllerManager:
  featureGates:
    CustomResourceValidation: true
```

The `zone` field tells the cloud-controller-manager in which zone it should mainly operate.
You can still create clusters in multiple availability zones, however, the cloud-controller-manager requires one "main" zone.
:warning: You always have to specify this field!

The `cloudControllerManager.featureGates` contains a map of explicitly enabled or disabled feature gates.
For production usage it's not recommend to use this field at all as you can enable alpha features or disable beta/stable features, potentially impacting the cluster stability.
If you don't want to configure anything for the `cloudControllerManager` simply omit the key in the YAML specification.

## Example `Shoot` manifest (one availability zone)

Please find below an example `Shoot` manifest for one availability zone:

```yaml
apiVersion: core.gardener.cloud/v1alpha1
kind: Shoot
metadata:
  name: johndoe-alicloud
  namespace: garden-dev
spec:
  cloudProfileName: alicloud
  region: eu-central-1
  secretBindingName: core-alicloud
  provider:
    type: alicloud
    infrastructureConfig:
      apiVersion: alicloud.provider.extensions.gardener.cloud/v1alpha1
      kind: InfrastructureConfig
      networks:
        vpc:
          cidr: 10.250.0.0/16
        zones:
        - name: eu-central-1a
          worker: 10.250.0.0/19
    controlPlaneConfig:
      apiVersion: alicloud.provider.extensions.gardener.cloud/v1alpha1
      kind: ControlPlaneConfig
      zone: eu-central-1a
    workers:
    - name: worker-xoluy
      machine:
        type: ecs.sn2ne.large
      minimum: 2
      maximum: 2
      volume:
        size: 50Gi
        type: cloud_efficiency
      zones:
      - eu-central-1a
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

## Example `Shoot` manifest (two availability zones)

Please find below an example `Shoot` manifest for two availability zones:

```yaml
apiVersion: core.gardener.cloud/v1alpha1
kind: Shoot
metadata:
  name: johndoe-alicloud
  namespace: garden-dev
spec:
  cloudProfileName: alicloud
  region: eu-central-1
  secretBindingName: core-alicloud
  provider:
    type: alicloud
    infrastructureConfig:
      apiVersion: alicloud.provider.extensions.gardener.cloud/v1alpha1
      kind: InfrastructureConfig
      networks:
        vpc:
          cidr: 10.250.0.0/16
        zones:
        - name: eu-central-1a
          worker: 10.250.0.0/26
        - name: eu-central-1b
          worker: 10.250.0.64/26
    controlPlaneConfig:
      apiVersion: alicloud.provider.extensions.gardener.cloud/v1alpha1
      kind: ControlPlaneConfig
      zone: eu-central-1a
    workers:
    - name: worker-xoluy
      machine:
        type: ecs.sn2ne.large
      minimum: 2
      maximum: 4
      volume:
        size: 50Gi
        type: cloud_efficiency
      zones:
      - eu-central-1a
      - eu-central-1b
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
