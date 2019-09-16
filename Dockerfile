#############      builder-base                             #############
FROM golang:1.13.0 AS builder-base

WORKDIR /go/src/github.com/gardener/gardener-extensions
COPY . .

RUN ./hack/install-requirements.sh

#############      builder                                  #############
FROM builder-base AS builder

ARG VERIFY=true

WORKDIR /go/src/github.com/gardener/gardener-extensions
COPY . .

RUN make VERIFY=$VERIFY all

#############      base                                     #############
FROM alpine:3.8 AS base

RUN apk add --update bash curl

WORKDIR /

#############      gardener-extension-hyper                 #############
FROM base AS gardener-extension-hyper

COPY controllers/provider-aws/charts /controllers/provider-aws/charts
COPY controllers/provider-azure/charts /controllers/provider-azure/charts
COPY controllers/provider-gcp/charts /controllers/provider-gcp/charts
COPY controllers/provider-openstack/charts /controllers/provider-openstack/charts
COPY controllers/provider-alicloud/charts /controllers/provider-alicloud/charts
COPY controllers/provider-packet/charts /controllers/provider-packet/charts

COPY controllers/extension-certificate-service/charts /controllers/extension-certificate-service/charts
COPY controllers/extension-shoot-dns-service/charts /controllers/extension-shoot-dns-service/charts

COPY controllers/networking-calico/charts /controllers/networking-calico/charts

COPY --from=builder /go/bin/gardener-extension-hyper /gardener-extension-hyper

ENTRYPOINT ["/gardener-extension-hyper"]
