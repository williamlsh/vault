FROM golang:1.13.1-alpine

WORKDIR /go/src/github.com/williamlsh/vault/

# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . .

# Enable Go Module
ENV GO111MODULE=on

RUN go install ./cmd/vaultd

FROM alpine:latest

COPY --from=0 /go/bin/vaultd .

ENTRYPOINT [ "./vaultd" ]

CMD ["-http-addr=:443", "-grpc-addr=:8080", "-prom-addr=:8081", "-tls-key=/testdata/server-key.pem", "-tls-cert=/testdata/server-cert.pem", "-pg-user=postgres", "-pg-password=postgres", "-pg-dbname=postgres", "-pg-host=localhost", "-pg-sslmode=disable", "-pg-port=5432"]

EXPOSE 443 8080-8081