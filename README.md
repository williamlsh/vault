# Vault

Vault provides hashing password and validating password and hash pair services. The development of Vault is under the book of [Go Programming Blueprints](https://www.goodreads.com/book/show/32902495-go-programming-blueprints---second-edition).

## Description

Vault is a simple microservice component for password hashing and validating service. It's mainly developed with [go-kit](https://gokit.io) and protocol buffer based [gRPC](https://grpc.io/).

Vault service is composed of three layers:

- Vault service
- Vault endpoints
  - hash endpoint
  - validate endpoint
- Vault transports
  - http transport
  - gRPC transport

Vault root is composed of three packages:

- vault: package defining three service layers.
- client: gRPC client used for vaultcli as entry to gRPC endpoint.
- cmd: executable commands.
  - vaultd: daemon executable for vault service, only ready for HTTP endpoint with command line.
  - vaultcli: client command line to consume vault service only with gRPC endpoint.

Vault uses only one middleware:

- `ratelimit` supported by golang.org/x/time/rate

## Installation

The application requires a working Go development environment.

```bash
go get github.com/williamzion/vault
```

## Usage

Run vaultd daemon first:

```bash
cd go list -f '{{.Dir}}' github.com/williamzion/vault/cmd/vaultd
```

```bash
go run main.go
```

- Consume HTTP endpoint.

  To hash password:

  ```bash
  curl -XPOST -d '{"password":"your_password"}' \
  http://localhost:8080/hash
  ```

  To validate password and hash pair:

  ```bash
  curl -XPOST -d '{"password":"your_password","hash":"previous_hash"}' \
  http://localhost:8080/validate
  ```

- Consume gRPC endpoint.

  Install vaultcli first:

  ```bash
  d go list -f '{{.Dir}}' github.com/williamzion/vault/cmd/vaultcli
  ```

  ```bash
  go install
  ```

  To hash password:

  ```bash
  vaultcli hash 'my_password'
  ```

  To validate password and hash pair:

  ```bash
  vaultcli validate 'previous_hash' 'my_password'
  ```

## Credits

- All credits to [matryer](https://github.com/matryer)
- [William](https://github.com/williamzion)
