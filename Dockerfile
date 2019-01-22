#############      builder       #############
FROM golang:1.11.4 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extensions
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install \
  -ldflags "-X github.com/gardener/gardener-extensions/controllers/os-coreos/pkg/version.Version=$(cat VERSION)" \
  ./controllers/os-coreos/cmd/gardener-extension-os-coreos

#############      os-coreos     #############
FROM alpine:3.8 AS os-coreos

RUN apk add --update bash curl

COPY --from=builder /go/bin/gardener-extension-os-coreos /gardener-extension-os-coreos

WORKDIR /

ENTRYPOINT ["/gardener-extension-os-coreos"]
