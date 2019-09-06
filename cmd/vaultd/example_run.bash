#!/bin/bash

export VAULTD_LOG_LEVEL=all

cd $(go list -f '{{.Dir}}' github.com/williamlsh/vault/cmd/vaultd) && go run \
  -race . \
  -http-addr=":8080" \
  -grpc-addr=":8081" \
  -debug-addr=":8082" \
  -tls-key="../../testdata/server-key.pem" \
  -tls-cert="../../testdata/server-cert.pem" \
  -pg-user="postgres" \
  -pg-password="postgres" \
  -pg-dbname="postgres" \
  -pg-host="localhost" \
  -pg-sslmode="disable" \
  -pg-port="5432"
