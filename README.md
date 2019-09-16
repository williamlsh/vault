# Vault

Vault provides bcrypt based password hashing and validating services.

## Table of contents

- [Layout](#Layout)
  - [Data Store](#Data-Store)
  - [Transport Security](#Transport-Security)
  - [Middleware](#Middleware)
  - [Client](#Client)
- [Installation](#Installation)
- [Usage](#Usage)
- [Docker Deployment](#Docker-Deployment)
- [CI Integration](#CI-Integration)
- [Credits](#Credits)
- [License](#License)

### Layout

Vault is a simple microservice component exposed a gRPC endpoint as well as a supplemental HTTP endpoint. It's mainly developed with [go-kit](https://gokit.io) and protocol buffers based [gRPC](https://grpc.io/).

#### Data Store

The `vault/pkg/store` package implements inner layer business logic with Postgres database. It exposes a `Store` interface which is highly decoupled. Data store implementation may not be really practical in such vault service case which is no more than just one `KeepSecret` method but the use of interface here is quite common and useful and could even be a trick to newbies.

#### Transport Security

Both gRPC and HTTP transports are implemented with **TLS encryption** and **JWT authentication**. HTTP with TLS could be easily tested in localhost environment.

To be noted here: the auth implementation between original gRPC and go-kit gRPC transport is a little different. Original gRPC uses `UnaryInterceptor` but not the case of go-kit due to the later one already had it integrated in transport layer.

#### Middleware

The service and endpoint layers both are implemented with middleware. Logging middleware for both and prometheus middleware for endpoint only but none for transport layer now.

#### Client

There are two kinds of clients, gRCP and HTTP clients corresponding to the two endpoints of vault service. The clients are not implemented customary but by use of go-kit client library in `vault/pkg/vaultransport`.

### Installation

The installation requires a Go development environment.

To enable go module(optional):

```bash
export GO111MODULE=on
```

To install `vaultd` service:

```bash
go get -u github.com/williamlsh/vault/cmd/vaultd
```

To install `vaultcli` client:

```bash
go get -u github.com/williamlsh/vault/cmd/vaultcli
```

### Usage

To run vaultd daemon:

```bash
vaultd \
  -http-addr=":443" \
  -grpc-addr=":8080" \
  -prom-addr=":8081" \ # prometheus metrics
  -tls-key="[KEY_FILE]" \ # private key
  -tls-cert="[CERT_FILE]" \ # certificate
  -pg-user="[PG_USER]" \
  -pg-password="[PG_PASS]" \
  -pg-dbname="[PG_DBNAME]" \
  -pg-host="[PG_HOST]" \
  -pg-sslmode="[PG_SSLMODE]" \
  -pg-port="[PG_PORT]"
```

To run gRPC client:

```bash
vaultcli \
  -server-name="[SERVER_NAME]" \ # localhost by default
  -tls-cert="[CERT_FILE]" \ # certificate
  -grpc-addr=":8080" \
  -method="[METHOD]" # hash or validate
```

To run HTTP client:

```bash
vaultcli \
  -http-addr=":443" \
  -method="[METHOD]" # hash or validate
```

To view Prometheus metrics at:

`https://localhost:8081/metrics`

### Docker Deployment

Vault can be easily deployed with Docker and Docker compose. There is already a latest docker image prebuilt on Docker hub registry: [williamofsino/vault](https://hub.docker.com/r/williamofsino/vault).

To run a single vault instance with Docker:

```bash
docker run -d \
  -e VAULTD_LOG_LEVEL=all \
  -p 8080-8081:8080-8081 \
  -p 443:443 \
  --mount source=./testdata/,target=/testdata/ \
  --name vault \
  --rm \
  williamofsino/vault:latest
```

To run entire service both vault and database with Docker compose:

```bash
docker-compose up -d
```

If you want to tear down the composed services, just run:

```bash
docker-compose down --volumes
```

### CI Integration

Vault is integrated with the following CI/CDs:

- Circle CI: for testing and building.
- Github Actions: for testing across all platforms.
- Docker hub: for building and pushing latest image.

### Credits

- [William](https://github.com/williamlsh)

### License

Under MIT license.
