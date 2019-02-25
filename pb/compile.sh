#!/usr/bin/env sh
protoc vault.proto --go_out=plugins=grpc:.