SHELL := /bin/bash

test: lint
	go test -v -race ./...

lint:
	golint ./...

up:
	docker-compose up -d

down:
	docker-compose down

deps-reset:
	git checkout -- go.mod

deps-upgrade:
	go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)

deps-cleancache:
	go clean -modcache