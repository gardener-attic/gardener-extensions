# [Gardener Extension for Packet provider](https://gardener.cloud)

[![Go Report Card](https://goreportcard.com/badge/github.com/gardener/gardener-extensions/controllers/provider-packet)](https://goreportcard.com/report/github.com/gardener/gardener-extensions/controllers/provider-packet)

Project Gardener implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service. Its main principle is to leverage Kubernetes concepts for all of its tasks.

Recently, most of the vendor specific logic has been developed [in-tree](https://github.com/gardener/gardener). However, the project has grown to a size where it is very hard to extend, maintain, and test. With [GEP-1](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md) we have proposed how the architecture can be changed in a way to support external controllers that contain their very own vendor specifics. This way, we can keep Gardener core clean and independent.

This controller operates on the `Infrastructure` resource in the `extensions.gardener.cloud/v1alpha1` API group. It manages those objects that are requesting an Packet infrastructure (`.spec.type=packet`):

```yaml
---
apiVersion: extensions.gardener.cloud/v1alpha1
kind: Infrastructure
metadata:
  name: infrastructure
spec:
  type: packet
  region: ewr1
  secretRef:
    name: cloudprovider
    namespace: shoot--foobar--packet
  providerConfig:
    apiVersion: packet.provider.extensions.gardener.cloud/v1alpha1
    kind: InfrastructureConfig
  sshPublicKey: ...

```

Please find [a concrete example](example/infrastructure.yaml) in the `example` folder.

After reconciliation the resulting data will be stored in the resource's `.status` field:

```yaml
...
status:
  ...
  providerStatus:
    apiVersion: packet.provider.extensions.gardener.cloud/v1alpha1
    kind: InfrastructureStatus
    ...
```

An example for a `ControllerRegistration` resource that can be used to register this controller to Gardener can be found [here](example/controller-registration.yaml).

Please find more information regarding the extensibility concepts and a detailed proposal [here](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md).

----

## How to start using or developing this extension controller locally

You can run the controller locally on your machine by executing `make start-provider-packet`.

Static code checks and tests can be executed by running `VERIFY=true make all`. We are using [dep](https://github.com/golang/dep) for Golang package dependency management and [Ginkgo](https://github.com/onsi/ginkgo)/[Gomega](https://github.com/onsi/gomega) for testing.

## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as [GitHub issues](https://github.com/gardener/gardener-extensions/issues) or join our [Slack channel #gardener](https://kubernetes.slack.com/messages/gardener) (please invite yourself to the Kubernetes workspace [here](http://slack.k8s.io)).

## Learn more!

Please find further resources about out project here:

* [Our landing page gardener.cloud](https://gardener.cloud/)
* ["Gardener, the Kubernetes Botanist" blog on kubernetes.io](https://kubernetes.io/blog/2018/05/17/gardener/)
* [GEP-1 (Gardener Enhancement Proposal) on extensibility](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)
