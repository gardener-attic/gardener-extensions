#############      builder                                  #############
FROM golang:1.11.5 AS builder

ARG VERIFY=true

WORKDIR /go/src/github.com/gardener/gardener-extensions
COPY . .

RUN ./hack/install-requirements.sh

RUN make VERIFY=$VERIFY all

#############      base                                     #############
FROM alpine:3.8 AS base

RUN apk add --update bash curl

WORKDIR /

#############      gardener-extension-os-coreos             #############
FROM base AS gardener-extension-os-coreos

COPY --from=builder /go/bin/gardener-extension-os-coreos /gardener-extension-os-coreos

ENTRYPOINT ["/gardener-extension-os-coreos"]

#############      gardener-extension-os-coreos-alicloud     #############
FROM base AS gardener-extension-os-coreos-alicloud

COPY --from=builder /go/bin/gardener-extension-os-coreos-alicloud /gardener-extension-os-coreos-alicloud

ENTRYPOINT ["/gardener-extension-os-coreos-alicloud"]

#############      gardener-extension-hyper                 #############
FROM base AS gardener-extension-hyper

COPY --from=builder /go/bin/gardener-extension-hyper /gardener-extension-hyper

ENTRYPOINT ["/gardener-extension-hyper"]
