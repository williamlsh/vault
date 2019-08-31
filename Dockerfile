# Reference: https://blog.container-solutions.com/faster-builds-in-docker-with-go-1-11

FROM golang:1.12-alpine AS build

# Install tools required for project
RUN apk add --no-cache bash ca-certificates git gcc g++ libc-dev

# These layers are only re-built when Go files are updated
COPY go.mod go.sum /go/src/github.com/williamlsh/vault/

WORKDIR /go/src/github.com/williamlsh/vault

# Enable Go Module
ENV GO111MODULE=on

# Install library dependencies
RUN go mod download

# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . /go/src/github.com/williamlsh/vault/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -a -tags vaultd -ldflags '-w -extldflags "-static"' ./cmd/vaultd

FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=build /go/bin/vaultd /bin/vaultd

ENTRYPOINT [ "/bin/vaultd" ]

EXPOSE 8080-8082