# Vault

Vault provides bcrypt based password hashing and validating services.

## Description

Vault is a simple microservice component exposed a gRPC endpoint as well as a supplemented HTTP endpoint. It's mainly developed with [go-kit](https://gokit.io) and protocol buffers based [gRPC](https://grpc.io/).

### Data store

The `vault/pkg/store` package implements inner layer business logic with Postgres database. It exposes a `Store` interface which is highly decoupled. Data store implementation may not be really practical in such vault service case which is no more than just one `KeepSecret` method but the use of interface here is quite common and useful and could even be a trick to newbies.

### Transport security

Since gRPC is the primary transport here, I only implemented gRPC transport with TLS encryption and JWT authentication. HTTP with TLS could be easily implemented but local test is not convenient.

To be noted here: the auth implementation between original gRPC and go-kit gRPC transport is a little different. Original gRPC uses `UnaryInterceptor` but not the case of go-kit due to the later one already had it integrated in transport layer.

### Client

There are two kinds of clients, gRCP and HTTP clients corresponding to the two endpoints of vault service. The clients are not implemented customary but by use of go-kit client library in `vault/pkg/vaultransport`.

## Installation

The installation requires a Go development environment.

To enable go module(optional):

```bash
export GO111MODULE=on
```

To install `vaultd` service:

```bash
go get -u github.com/williamzion/vault/cmd/vaultd
```

To install `vaultcli` client:

```bash
go get -u github.com/williamzion/vault/cmd/vaultcli
```

## Usage

To run vaultd daemon:

```bash
vaultd \
  -http-addr=":8080" \
  -grpc-addr=":8081" \
  -key-file="[KEY_FILE]" \ # private key
  -cert-file="[CERT_FILE]" \ # certificate
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
  -server-name="[SERVER_NAME]" \ # server name in csr file
  -cert-file="[CERT_FILE]" \ # certificate
  -grpc-addr=":8081" \
  -method="[METHOD]" # hash or validate
```

To run HTTP client:

```bash
vaultcli \
  -http-addr=":8080" \
  -method="[METHOD]" # hash or validate
```

## Credits

- [William](https://github.com/williamzion)

## License

Under MIT license.
