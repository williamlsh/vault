#!/usr/bin/env bash

export VAULTD_LOG_LEVEL=all
cd $(go list -f '{{.Dir}}' github.com/williamzion/vault/cmd/vaultd) && go run \
  -race . \
  -http-addr=":8080" \
  -grpc-addr=":8081"
