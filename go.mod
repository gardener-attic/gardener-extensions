module github.com/gardener/gardener-extensions

go 1.12

require (
	cloud.google.com/go v0.43.0
	github.com/Azure/azure-sdk-for-go v32.6.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.7.0
	github.com/Azure/go-autorest/autorest/azure/auth v0.3.0
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.4.2
	github.com/aliyun/alibaba-cloud-sdk-go v0.0.0-20190723075400-e63e3f9dd712
	github.com/aliyun/aliyun-oss-go-sdk v2.0.1+incompatible
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30
	github.com/aws/aws-sdk-go v1.21.10
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/gardener/external-dns-management v0.0.0-20190722114702-f6b12f6e4b43 // indirect
	github.com/gardener/gardener v0.0.0-20191029090635-befe718e1567
	github.com/gardener/gardener-resource-manager v0.0.0-20190828115855-7ceeb3021993
	github.com/gardener/machine-controller-manager v0.0.0-20190606071036-119056ee3fdd
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/gobuffalo/logger v1.0.1 // indirect
	github.com/gobuffalo/packd v0.3.0 // indirect
	github.com/gobuffalo/packr v1.25.0
	github.com/gobuffalo/packr/v2 v2.1.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/mock v1.3.1
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gophercloud/gophercloud v0.3.0
	github.com/gophercloud/utils v0.0.0-20190527093828-25f1b77b8c03
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/huandu/xstrings v1.2.0
	github.com/jetstack/cert-manager v0.6.2
	github.com/karrick/godirwalk v1.10.12 // indirect
	github.com/nwaples/rardecode v1.0.0 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/packethost/packngo v0.0.0-20181217122008-b3b45f1b4979
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.7.0
	gopkg.in/ini.v1 v1.44.0 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20191004102349-159aefb8556b
	k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery v0.0.0-20191004074956-c5d2f014d689
	k8s.io/apiserver v0.0.0-20191010014313-3893be10d307
	k8s.io/autoscaler v0.0.0-20190425094807-60ef58f770d8
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.0.0-20190713022532-93d7507fc8ff
	k8s.io/component-base v0.0.0-20190816222507-f3799749b6b7
	k8s.io/gengo v0.0.0-20190826232639-a874a240740c
	k8s.io/helm v2.14.2+incompatible
	k8s.io/klog v0.3.3
	k8s.io/kube-aggregator v0.0.0-20191004104030-d9d5f0cc7532
	k8s.io/kube-openapi v0.0.0-20190722073852-5e22f3d471e6 // indirect
	k8s.io/kubelet v0.0.0-20190314002251-f6da02f58325
	k8s.io/utils v0.0.0-20190712204705-3dccf664f023 // indirect
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab //kubernetes-1.14.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed // kubernetes-1.14.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1 // kubernetes-1.14.0
	k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible // kubernetes-1.14.0
)
