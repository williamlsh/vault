#!/bin/bash

export VAULTCLI_LOG_LEVEL="all"

# Run gRPC client.
cd $(go list -f '{{.Dir}}' github.com/williamlsh/vault/cmd/vaultcli) && go run \
  -race . \
  -server-name="example.com" \
  -tls-cert="../../testdata/server-cert.pem" \
  -grpc-addr=":8081" \
  -method="hash"

# Run HTTP client.
cd $(go list -f '{{.Dir}}' github.com/williamlsh/vault/cmd/vaultcli) && go run \
  -race . \
  -http-addr=":8080" \
  -method="hash"
