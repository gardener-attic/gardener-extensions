module github.com/gardener/gardener-extensions

go 1.13

require (
	cloud.google.com/go/storage v1.0.0
	github.com/Azure/azure-sdk-for-go v32.6.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.7.0
	github.com/Azure/go-autorest/autorest/azure/auth v0.3.0
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/ahmetb/gen-crd-api-reference-docs v0.1.5
	github.com/aliyun/alibaba-cloud-sdk-go v1.60.340
	github.com/aliyun/aliyun-oss-go-sdk v2.0.1+incompatible
	github.com/aws/aws-sdk-go v1.21.10
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/gardener/etcd-druid v0.0.0-20191206092814-9c2731e221e3
	github.com/gardener/gardener v0.34.0
	github.com/gardener/gardener-resource-manager v0.8.1
	github.com/gardener/machine-controller-manager v0.25.1-0.20200115123605-0510de7ddfca // master
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/packr v1.25.0
	github.com/gobuffalo/packr/v2 v2.1.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/mock v1.3.1
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gophercloud/gophercloud v0.7.0
	github.com/gophercloud/utils v0.0.0-20190527093828-25f1b77b8c03
	github.com/huandu/xstrings v1.2.1
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/nwaples/rardecode v1.0.0 // indirect
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/packethost/packngo v0.0.0-20181217122008-b3b45f1b4979
	github.com/pierrec/lz4 v2.4.1+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	go.uber.org/zap v1.13.0
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
	gomodules.xyz/jsonpatch/v2 v2.0.1
	google.golang.org/api v0.14.0
	google.golang.org/grpc v1.23.1 // indirect
	gopkg.in/ini.v1 v1.51.1 // indirect
	gopkg.in/yaml.v2 v2.2.7
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/apiserver v0.0.0-20191010014313-3893be10d307
	k8s.io/autoscaler v0.0.0-20190805135949-100e91ba756e
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269
	k8s.io/component-base v0.0.0-20190918160511-547f6c5d7090
	k8s.io/gengo v0.0.0-20190826232639-a874a240740c
	k8s.io/helm v2.16.1+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.0.0-20191004104030-d9d5f0cc7532
	k8s.io/kubelet v0.0.0-20190918162654-250a1838aa2c
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8
	github.com/gardener/etcd-druid => github.com/georgekuruvillak/etcd-druid-1 v0.0.0-20200106104847-f1d9b72827af
	github.com/gardener/external-dns-management => github.com/gardener/external-dns-management v0.0.0-20190927090840-6659f5a46d13
	github.com/gardener/gardener => github.com/georgekuruvillak/gardener v0.0.0-20200123100617-cb8b9668ce75
	k8s.io/api => k8s.io/api v0.0.0-20190918155943-95b840bb6a1f // kubernetes-1.16.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783 // kubernetes-1.16.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655 // kubernetes-1.16.0
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918160949-bfa5e2e684ad // kubernetes-1.16.0
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90 // kubernetes-1.16.0
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269 // kubernetes-1.16.0
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190918160511-547f6c5d7090 // kubernetes-1.16.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190918161219-8c8f079fddc3 // kubernetes-1.16.0
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190918162654-250a1838aa2c // kubernetes-1.16.0
)
