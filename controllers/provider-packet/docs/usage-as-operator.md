# Using the Packet provider extension with Gardener as operator

The [`core.gardener.cloud/v1alpha1.CloudProfile` resource](https://github.com/gardener/gardener/blob/master/example/30-cloudprofile.yaml) declares a `providerConfig` field that is meant to contain provider-specific configuration.

In this document we are describing how this configuration looks like for Packet and provide an example `CloudProfile` manifest with minimal configuration that you can use to allow creating Packet shoot clusters.

## `CloudProfileConfig`

The cloud profile configuration contains information about the real machine image IDs in the Packet environment (IDs).
You have to map every version that you specify in `.spec.machineImages[].versions` here such that the Packet extension knows the ID for every version you want to offer.

An example `CloudProfileConfig` for the Packet extension looks as follows:

```yaml
apiVersion: packet.provider.extensions.gardener.cloud/v1alpha1
kind: CloudProfileConfig
machineImages:
- name: coreos
  versions:
  - version: 2135.6.0
    id: coreos-2135.6.0-id
```

## Example `CloudProfile` manifest

Please find below an example `CloudProfile` manifest:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: CloudProfile
metadata:
  name: packet
spec:
  type: packet
  kubernetes:
    versions:
    - version: 1.16.1
    - version: 1.16.0
      expirationDate: "2020-04-05T01:02:03Z"
  machineImages:
  - name: coreos
    versions:
    - version: 2135.6.0
  machineTypes:
  - name: t1.small
    cpu: "4"
    gpu: "0"
    memory: 8Gi
    usable: true
  volumeTypes:
  - name: storage_1
    class: standard
    usable: true
  - name: storage_2
    class: performance
    usable: true
  regions:
  - name: EWR1
  providerConfig:
    apiVersion: packet.provider.extensions.gardener.cloud/v1alpha1
    kind: CloudProfileConfig
    machineImages:
    - name: coreos
      versions:
      - version: 2135.6.0
        id: coreos-2135.6.0-id
```
