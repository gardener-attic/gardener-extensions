#############      builder-base                             #############
FROM golang:1.11.5 AS builder-base

COPY ./hack/install-requirements.sh /install-requirements.sh

RUN /install-requirements.sh

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

COPY --from=builder /go/bin/gardener-extension-hyper /gardener-extension-hyper

ENTRYPOINT ["/gardener-extension-hyper"]
