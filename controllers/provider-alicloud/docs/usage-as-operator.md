# Using the Alicloud provider extension with Gardener as operator

The [`core.gardener.cloud/v1alpha1.CloudProfile` resource](https://github.com/gardener/gardener/blob/master/example/30-cloudprofile.yaml) declares a `providerConfig` field that is meant to contain provider-specific configuration.

In this document we are describing how this configuration looks like for Alicloud and provide an example `CloudProfile` manifest with minimal configuration that you can use to allow creating Alicloud shoot clusters.

In addition, this document also describes how to enable the use of customized machine images for Alicloud.

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

## Enable customized machine images for the Alicloud extension

Customized machine images can be created for an Alicloud account and shared with other Alicloud accounts. The same customized machine image has different image ID in different regions on Alicloud. Administrators/Operators need to explicitly declare them per imageID per region as below:

```yaml
machineImages:
- name: customized_coreos
  regions:
  - imageID: <image_id_in_eu_central_1>
    region: eu-central-1
  - imageID: <image_id_in_cn_shanghai>
    region: cn-shanghai
  ...
  version: 2191.4.1
...
```

End-users have to have the permission to use the customized image from its creator Alicloud account. To enable end-users to use customized images, the images are shared from Alicloud account of Seed operator with end-users' Alicloud accounts. Administrators/Operators need to explicitly provide Seed operator's Alicloud account access credentials (base64 encoded) as below:

```yaml
machineImageOwnerSecret:
  name: machine-image-owner
  accessKeyID: <base64_encoded_access_key_id>
  accessKeySecret: <base64_encoded_access_key_secret>
```

As a result, a Secret named `machine-image-owner` by default will be created in namespace of Alicloud provider extension.

## Example `ControllerRegistration` manifest for enabling customized machine images

```yaml
apiVersion: core.gardener.cloud/v1alpha1
kind: ControllerRegistration
metadata:
  name: extension-provider-alicloud
spec:
  deployment:
    type: helm
    providerConfig:
      chart: |
        H4sIFAAAAAAA/yk...
      values:
        config:
          machineImageOwnerSecret:
            accessKeyID: <base64_encoded_access_key_id>
            accessKeySecret: <base64_encoded_access_key_secret>
          machineImages:
          - name: customized_coreos
            regions:
            - imageID: <image_id_in_eu_central_1>
              region: eu-central-1
            - imageID: <image_id_in_cn_shanghai>
              region: cn-shanghai
            ...
            version: 2191.4.1
          ...
        resources:
          limits:
            cpu: 500m
            memory: 1Gi
          requests:
            memory: 128Mi
  resources:
  - kind: BackupBucket
    type: alicloud
  - kind: BackupEntry
    type: alicloud
  - kind: ControlPlane
    type: alicloud
  - kind: Infrastructure
    type: alicloud
  - kind: Worker
    type: alicloud
```
