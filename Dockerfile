#############      builder-base                             #############
FROM golang:1.13.4 AS builder-base

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
FROM alpine:3.10.3 AS base

RUN apk add --update bash curl

WORKDIR /

#############      gardener-extension-hyper                 #############
FROM base AS gardener-extension-hyper

COPY controllers/provider-aws/charts /controllers/provider-aws/charts
COPY controllers/provider-alicloud/charts /controllers/provider-alicloud/charts

COPY --from=builder /go/bin/gardener-extension-hyper /gardener-extension-hyper

ENTRYPOINT ["/gardener-extension-hyper"]
