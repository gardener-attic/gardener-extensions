# Using the Alicloud provider extension with Gardener as operator

The [`core.gardener.cloud/v1alpha1.CloudProfile` resource](https://github.com/gardener/gardener/blob/master/example/30-cloudprofile.yaml) declares a `providerConfig` field that is meant to contain provider-specific configuration.

In this document we are describing how this configuration looks like for Alicloud and provide an example `CloudProfile` manifest with minimal configuration that you can use to allow creating Alicloud shoot clusters.

## `CloudProfileConfig`

The cloud profile configuration contains information about the real machine image IDs in the Alicloud environment (AMIs).
You have to map every version that you specify in `.spec.machineImages[].versions` here such that the Alicloud extension knows the AMI for every version you want to offer.

An example `CloudProfileConfig` for the Alicloud extension looks as follows:

```yaml
apiVersion: alicloud.provider.extensions.gardener.cloud/v1alpha1
kind: CloudProfileConfig
machineImages:
- name: coreos
  version: 2023.4.0
  regions:
  - name: eu-central-1
    id: coreos_2023_4_0_64_30G_alibase_20190319.vhd
```

## Example `CloudProfile` manifest

Please find below an example `CloudProfile` manifest:

```yaml
apiVersion: core.gardener.cloud/v1alpha1
kind: CloudProfile
metadata:
  name: alicloud
spec:
  type: alicloud
  kubernetes:
    versions:
    - version: 1.16.1
    - version: 1.16.0
      expirationDate: "2020-04-05T01:02:03Z"
  machineImages:
  - name: coreos
    versions:
    - version: 2023.4.0
  machineTypes:
  - name: ecs.sn2ne.large
    cpu: "2"
    gpu: "0"
    memory: 8Gi
  volumeTypes:
  - name: cloud_efficiency
    class: standard
  - name: cloud_ssd
    class: premium
  regions:
  - name: eu-central-1
    zones:
    - name: eu-central-1a
    - name: eu-central-1b
  providerConfig:
    apiVersion: alicloud.provider.extensions.gardener.cloud/v1alpha1
    kind: CloudProfileConfig
    machineImages:
    - name: coreos
      version: 2023.4.0
      regions:
      - name: eu-central-1
        id: coreos_2023_4_0_64_30G_alibase_20190319.vhd
```
