# [Gardener Extensions](https://gardener.cloud)

[![CI Build status](https://concourse.ci.infra.gardener.cloud/api/v1/teams/gardener/pipelines/gardener-extensions-master/jobs/master-head-update-job/badge)](https://concourse.ci.infra.gardener.cloud/teams/gardener/pipelines/gardener-extensions-master/jobs/master-head-update-job)

Project Gardener implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service. Its main principle is to leverage Kubernetes concepts for all of its tasks.

Recently, most of the vendor specific logic has been developed [in-tree](https://github.com/gardener/gardener). However, the project has grown to a size where it is very hard to extend, maintain, and test. With [GEP-1](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md) we have proposed how the architecture can be changed in a way to support external controllers that contain their very own vendor specifics. This way, we can keep Gardener core clean and independent.

This repository contains utilities functions and common libraries meant to ease writing the actual extension controllers.

:warning: In order to gain synergies from commonly developed packages we are maintaining all extensions in this repository for a short period of time. Once these libraries have matured we will move out the extension controllers into their own dedicated repositories, e.g. https://github.com/gardener/gardener-extension-os-coreos.

## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as [GitHub issues](https://github.com/gardener/gardener-extensions/gardener-extension-os-coreos/issues) or join our [Slack channel #gardener](https://kubernetes.slack.com/messages/gardener) (please invite yourself to the Kubernetes workspace [here](http://slack.k8s.io)).

## Learn more!

Please find further resources about out project here:

* [Our landing page gardener.cloud](https://gardener.cloud/)
* ["Gardener, the Kubernetes Botanist" blog on kubernetes.io](https://kubernetes.io/blog/2018/05/17/gardener/)
* [GEP-1 (Gardener Enhancement Proposal) on extensibility](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)
