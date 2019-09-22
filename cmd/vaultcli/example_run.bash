#!/bin/bash

export GO111MODULE=off
export VAULTCLI_LOG_LEVEL="all"

# Run gRPC client.
cd $(go list -f '{{.Dir}}' github.com/williamlsh/vault/cmd/vaultcli) && go run \
  -race . \
  -server-name="localhost" \
  -tls-cert="../../testdata/server-cert.pem" \
  -grpc-addr=":8080" \
  -method="hash" \
  -zipkin-reporter-url="" \
  -zipkin-endpoint-port="" \
  -lightstep-token="" \
  -appdash-addr=""

# Run HTTP client.
cd $(go list -f '{{.Dir}}' github.com/williamlsh/vault/cmd/vaultcli) && go run \
  -race . \
  -http-addr="https://localhost:443" \
  -method="hash" \
  -zipkin-reporter-url="" \
  -zipkin-endpoint-port="" \
  -lightstep-token="" \
  -appdash-addr=""
