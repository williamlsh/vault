#!/bin/bash

export GO111MODULE=off
export VAULTD_LOG_LEVEL=all

cd $(go list -f '{{.Dir}}' github.com/williamlsh/vault/cmd/vaultd) && go run \
  -race . \
  -http-addr=":443" \
  -grpc-addr=":8080" \
  -prom-addr=":8081" \
  -tls-key="../../testdata/server-key.pem" \
  -tls-cert="../../testdata/server-cert.pem" \
  -pg-user="postgres" \
  -pg-password="postgres" \
  -pg-dbname="postgres" \
  -pg-host="localhost" \
  -pg-sslmode="disable" \
  -pg-port="5432" \
  -zipkin-reporter-url="http://zipkin:9411/api/v2/spans" \
  -zipkin-endpoint-port="8082" \
  -lightstep-token="Rcd6X6kZcBfEG2Kvq4lXxuFqWhyt2cHJFWBpOQyAJQpdvBwyc2D9usKQaL/MZCkXSs5eKMNoCRa3dQOJh8tBB999aFTAEy4kwB1YF6p1" \
  -appdash-addr=""
