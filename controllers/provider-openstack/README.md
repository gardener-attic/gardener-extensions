# [Gardener Extension for OpenStack provider](https://gardener.cloud)

[![Go Report Card](https://goreportcard.com/badge/github.com/gardener/gardener-extensions/controllers/provider-openstack)](https://goreportcard.com/report/github.com/gardener/gardener-extensions/controllers/provider-openstack)

Project Gardener implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service. Its main principle is to leverage Kubernetes concepts for all of its tasks.

Recently, most of the vendor specific logic has been developed [in-tree](https://github.com/gardener/gardener). However, the project has grown to a size where it is very hard to extend, maintain, and test. With [GEP-1](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md) we have proposed how the architecture can be changed in a way to support external controllers that contain their very own vendor specifics. This way, we can keep Gardener core clean and independent.

< needs-to-be-implemented >

An example for a `ControllerRegistration` resource that can be used to register this controller to Gardener can be found [here](example/controller-registration.yaml).

Please find more information regarding the extensibility concepts and a detailed proposal [here](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md).

----

## How to start using or developing this extension controller locally

You can run the controller locally on your machine by executing `make start-provider-openstack`.

Static code checks and tests can be executed by running `VERIFY=true make all`. We are using [dep](https://github.com/golang/dep) for Golang package dependency management and [Ginkgo](https://github.com/onsi/ginkgo)/[Gomega](https://github.com/onsi/gomega) for testing.

## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as [GitHub issues](https://github.com/gardener/gardener-extensions/issues) or join our [Slack channel #gardener](https://kubernetes.slack.com/messages/gardener) (please invite yourself to the Kubernetes workspace [here](http://slack.k8s.io)).

## Learn more!

Please find further resources about out project here:

* [Our landing page gardener.cloud](https://gardener.cloud/)
* ["Gardener, the Kubernetes Botanist" blog on kubernetes.io](https://kubernetes.io/blog/2018/05/17/gardener/)
* [GEP-1 (Gardener Enhancement Proposal) on extensibility](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)


## The Controllers for Openstack

### Control Plane

The control plane controller uses the following provider config for the control plane extension object:

```yaml
apiVersion: openstack.provider.extensions.gardener.cloud/v1alpha1
kind: ControlPlaneConfig
metadata:
  name:
  namespace:
  
# the name of the load balancer provider, e.g. haproxy
loadBalancerProvider: <string>
# list of load balancer classes
loadBalancerClasses:
- name: <string>
  
  # ID of subnet in the floating pool network (optional)
  floatingSubnetID: <string>
  # ID of the tenant local subnet to be used for (private) load balancer deployment (optional)
  subnetID: <string>
  

#  a map of enabled/disabled feature gates for the controller manager
cloudControllerManager:
  <feature>: <bool>
```

The network id of the provider network is taken from the infrastructure status.
In the shoot it is configured by its name (the floating pool name). By the infrastructure
part it is mapped to a its id which is configured for the router as provider network.
The id of the configured provider network is then finally exported in the status.

If the profile configures load balancer classes with dedicated floating subnet ids, these
subnets must be on the selected floating network. Those classes, or manually maintained
classes are then put into the soot manifest and finnaly arrive here using the 
structure shown above.

If the router supports access to multiple provider networks, therevmay be multiple classes
configured with different provider networks. If not present a required provider network
in a class will be default from the network used for the chosen flpoating pool.

### Infrastructure

### Worker