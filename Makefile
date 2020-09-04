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
	go get -v -t -u ./...

deps-cleancache:
	go clean -modcache