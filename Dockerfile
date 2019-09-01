# Reference: https://blog.container-solutions.com/faster-builds-in-docker-with-go-1-11

FROM golang:1.12-alpine AS builder

# Install tools required for project
RUN apk --no-cache add ca-certificates git gcc g++ libc-dev

WORKDIR /go/src/github.com/williamlsh/vault/

# These layers are only re-built when Go files are updated
COPY go.mod go.sum /

# Enable Go Module
ENV GO111MODULE=on

# Install library dependencies
RUN go mod download

# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -a ./cmd/vaultd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/bin/vaultd /bin/vaultd

COPY --from=builder /go/src/github.com/williamlsh/vault/testdata/ /testdata/

ENTRYPOINT [ "/bin/vaultd" ]

CMD ["-http-addr=:8080", "-grpc-addr=:8081", "-debug-addr=:8082", "-tls-key=/testdata/server-key.pem", "-tls-cert=/testdata/server-cert.pem", "-pg-user=postgres", "-pg-password=postgres", "-pg-dbname=postgres", "-pg-host=localhost", "-pg-sslmode=disable", "-pg-port=5432"]

EXPOSE 8080-8082