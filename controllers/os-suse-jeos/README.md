# [Gardener Extension for SUSE JeOS](https://gardener.cloud)

[![Go Report Card](https://goreportcard.com/badge/github.com/gardener/gardener-extensions/controllers/os-suse-jeos)](https://goreportcard.com/report/github.com/gardener/gardener-extensions/controllers/os-suse-jeos)

This controller operates on the [`OperatingSystemConfig`](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md#cloud-config-user-data-for-bootstrapping-machines) resource in the `extensions.gardener.cloud/v1alpha1` API group. It manages those objects that are requesting [SUSE JeOS](https://www.suse.com/products/server/jeos/) configuration (`.spec.type=suse-jeos`):

```yaml
---
apiVersion: extensions.gardener.cloud/v1alpha1
kind: OperatingSystemConfig
metadata:
  name: pool-01-original
  namespace: default
spec:
  type: suse-jeos
  units:
    ...
  files:
    ...
```

Please find [a concrete example](example/operatingsystemconfig.yaml) in the `example` folder.

After reconciliation the resulting data will be stored in a secret within the same namespace (as the config itself might contain confidential data). The name of the secret will be written into the resource's `.status` field:

```yaml
...
status:
  ...
  cloudConfig:
    secretRef:
      name: osc-result-pool-01-original
      namespace: default
  command: /usr/bin/env bash <path>
  units:
  - docker-monitor.service
  - kubelet-monitor.service
  - kubelet.service
```

The secret has one data key `cloud_config` that stores the generation.

An example for a `ControllerRegistration` resource that can be used to register this controller to Gardener can be found [here](example/controller-registration.yaml).

This controller is implemented using the [`oscommon`](https://github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig/oscommon/README.md) library for operating system configuration controllers.

Please find more information regarding the extensibility concepts and a detailed proposal [here](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md).

----

## How to start using or developing this extension controller locally

You can run the controller locally on your machine by executing `make start-os-suse-jeos`. Please make sure to have the kubeconfig to the cluster you want to connect to ready in the `./dev/kubeconfig` file.
Static code checks and tests can be executed by running `VERIFY=true make all`. We are using [dep](https://github.com/golang/dep) for Golang package dependency management and [Ginkgo](https://github.com/onsi/ginkgo)/[Gomega](https://github.com/onsi/gomega) for testing.

## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as [GitHub issues](https://github.com/gardener/gardener-extensions/issues) or join our [Slack channel #gardener](https://kubernetes.slack.com/messages/gardener) (please invite yourself to the Kubernetes workspace [here](http://slack.k8s.io)).

## Learn more!

Please find further resources about out project here:

* [Our landing page gardener.cloud](https://gardener.cloud/)
* ["Gardener, the Kubernetes Botanist" blog on kubernetes.io](https://kubernetes.io/blog/2018/05/17/gardener/)
* [GEP-1 (Gardener Enhancement Proposal) on extensibility](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)
