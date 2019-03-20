# [Gardener Extension for certificate services](https://gardener.cloud)

[![Go Report Card](https://goreportcard.com/badge/github.com/gardener/gardener-extensions/controllers/extension-certificate-service)](https://goreportcard.com/report/github.com/gardener/gardener-extensions/controllers/extension-certificate-service)

Project Gardener implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service. Its main principle is to leverage Kubernetes concepts for all of its tasks.

Recently, most of the vendor specific logic has been developed [in-tree](https://github.com/gardener/gardener). However, the project has grown to a size where it is very hard to extend, maintain, and test. With [GEP-1](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md) we have proposed how the architecture can be changed in a way to support external controllers that contain their very own vendor specifics. This way, we can keep Gardener core clean and independent.

## Components
Certificate services for Shoot clusters are provided by the composition of two components:

- [Cert-Manager](https://github.com/jetstack/cert-manager)
- [Cert-Broker](https://github.com/gardener/cert-broker)

## Configuration
Example configuration which is reconciled by this extension controller:
```yaml
apiVersion: certificate-service.extensions.config.gardener.cloud/v1alpha1
kind: Configuration
metadata:
  name: certificate-service
  namespace: default
spec:
  issuerName: lets-encrypt
  namespaceRef: default
  resourceNamespace: garden
  acme:
    email: john.doe@example.com
    server: https://acme-v02.api.letsencrypt.org/directory
  providers:
    clouddns:
    - name: clouddns-prod
      domains:
      - example.io
      project: project_id
      serviceAccount: |
        {
        "type": "service_account",
        "project_id": "demo-project"
        }
    route53:
    - name: route53-prod
      domains:
      - example.org
      region: us-east-1
      accessKeyID: your-accessKeyID
      secretAccessKey: your-secretAccessKey
```

The extension controller will create an instance of Cert-Manager as well as a [ClusterIsser](https://docs.cert-manager.io/en/latest/reference/clusterissuers.html) with the information provided above.
(Cert-Manager is responsible for managing certificate requests / renewals within the Seed cluster for configured Shoot domains)

## Extension-Resources
Besides the [configuration](#Configuration), operations also happen on [`Extension`](https://github.com/gardener/gardener/blob/master/pkg/apis/extensions/v1alpha1/types_extension.go) resources in the `extensions.gardener.cloud/v1alpha1` API group of type `.spec.type=certificate-service`:

```yaml
apiVersion: extensions.gardener.cloud/v1alpha1
kind: Extension
metadata:
  name: "extension-certificate-service"
  namespace: shoot--project--abc
spec:
  type: certificate-service
```

During reconciliation an instance of Cert-Broker is created in the very same namespace the `Extension` resource belongs to (Shoot namespace). Cert-Broker relays certificate requests from the Shoot cluster to the Cert-Manager in the Seed and puts back generated certificates. Events are propagated as well.

For security reasons each Shoot cluster must be restricted to only order certificates for domains it owns. The owning domain is extracted from the respective [`Cluster`](https://github.com/gardener/gardener/blob/master/pkg/apis/extensions/v1alpha1/types_cluster.go) resource.

## Kubeconfig for Shoot clusters

* **cert-broker**: Created with the `ca` secret from the Shoot's namespace in the Seed. This Kubeconfig is required by Cert-Broker to watch `Ingress` objects, create `Secrets` and `Events`.
  
* **gardener**: Used to apply `RBAC` manifests for Cert-Broker to the Shoot cluster. The `gardener` secret is already provided by the `Gardener-Controller-Manager`.


Please find more information regarding the extensibility concepts and a detailed proposal [here](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md).

----

## How to start using or developing this extension controller locally

You can run the controller locally on your machine by executing `make start-certificate-service`. 

Please make sure:
- `$KUBECONFIG` env is set accordingly.
- `Extension` CRD + `Cluster` CRD are applied.
- `ca` and `gardener` secrets for the Shoot cluster exists.
- [Configuration](#Configuration) is in place.

Static code checks and tests can be executed by running `VERIFY=true make all`. We are using [dep](https://github.com/golang/dep) for Golang package dependency management and [Ginkgo](https://github.com/onsi/ginkgo)/[Gomega](https://github.com/onsi/gomega) for testing.

## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as [GitHub issues](https://github.com/gardener/gardener-extensions/issues) or join our [Slack channel #gardener](https://kubernetes.slack.com/messages/gardener) (please invite yourself to the Kubernetes workspace [here](http://slack.k8s.io)).

## Learn more!

Please find further resources about out project here:

* [Our landing page gardener.cloud](https://gardener.cloud/)
* ["Gardener, the Kubernetes Botanist" blog on kubernetes.io](https://kubernetes.io/blog/2018/05/17/gardener/)
* [GEP-1 (Gardener Enhancement Proposal) on extensibility](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)
