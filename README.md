# [Gardener Extensions](https://gardener.cloud)

![Gardener Extensions Logo](logo/gardener-extension-180px.png)

[![CI Build status](https://concourse.ci.infra.gardener.cloud/api/v1/teams/gardener/pipelines/gardener-extensions-master/jobs/master-head-update-job/badge)](https://concourse.ci.infra.gardener.cloud/teams/gardener/pipelines/gardener-extensions-master/jobs/master-head-update-job)

Project Gardener implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service. Its main principle is to leverage Kubernetes concepts for all of its tasks.

Recently, most of the vendor specific logic has been developed [in-tree](https://github.com/gardener/gardener). However, the project has grown to a size where it is very hard to extend, maintain, and test. With [GEP-1](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md) we have proposed how the architecture can be changed in a way to support external controllers that contain their very own vendor specifics. This way, we can keep Gardener core clean and independent.

This repository contains utilities functions and common libraries meant to ease writing the actual extension controllers.
Please consult https://github.com/gardener/gardener/tree/master/docs/extensions to get more information about the extension contracts.

## Known Extension Implementations

Check out these repositories for implementations of the Gardener Extension contracts:

- [provider-alicloud](https://github.com/gardener/gardener-extension-provider-alicloud)
- [provider-aws](https://github.com/gardener/gardener-extension-provider-aws)
- [provider-azure](https://github.com/gardener/gardener-extension-provider-azure)
- [provider-gcp](https://github.com/gardener/gardener-extension-provider-gcp)
- [provider-metal](https://github.com/metal-pod/gardener-extension-provider-metal)
- [provider-openstack](https://github.com/gardener/gardener-extension-provider-openstack)
- [provider-packet](https://github.com/gardener/gardener-extension-provider-packet)
- [provider-vsphere](https://github.com/gardener/gardener-extension-provider-vsphere)
- [dns-external](https://github.com/gardener/external-dns-management)
- [os-coreos](https://github.com/gardener/gardener-extension-os-coreos)
- [os-coreos-alicloud](https://github.com/gardener/gardener-extension-os-coreos-alicloud)
- [os-metal](https://github.com/metal-pod/os-metal-extension)
- [os-ubuntu](https://github.com/gardener/gardener-extension-os-ubuntu)
- [os-ubuntu-alicloud](https://github.com/gardener/gardener-extension-os-ubuntu-alicloud)
- [os-suse-jeos](https://github.com/gardener/gardener-extension-os-suse-jeos)
- [networking-calico](https://github.com/gardener/gardener-extension-networking-calico)
- [networking-cilium](https://github.com/gardener/gardener-extension-networking-cilium)
- [shoot-cert-service](https://github.com/gardener/gardener-extension-shoot-cert-service)
- [shoot-dns-service](https://github.com/gardener/gardener-extension-shoot-dns-service)

If you implemented a new extension, please feel free to add it to this list!


## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as [GitHub issues](https://github.com/gardener/gardener-extensions/gardener-extensions/issues) or join our [Slack channel #gardener](https://kubernetes.slack.com/messages/gardener) (please invite yourself to the Kubernetes workspace [here](http://slack.k8s.io)).

## Learn more!

Please find further resources about out project here:

* [Our landing page gardener.cloud](https://gardener.cloud/)
* ["Gardener, the Kubernetes Botanist" blog on kubernetes.io](https://kubernetes.io/blog/2018/05/17/gardener/)
* ["Gardener Project Update" blog on kubernetes.io](https://kubernetes.io/blog/2019/12/02/gardener-project-update/)
* [GEP-1 (Gardener Enhancement Proposal) on extensibility](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)
* [Extensibility API documentation](https://github.com/gardener/gardener/tree/master/docs/extensions)
