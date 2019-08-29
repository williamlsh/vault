# Reference: https://blog.container-solutions.com/faster-builds-in-docker-with-go-1-11

FROM golang:latest-alpine AS build_base

WORKDIR $GOPATH/github.com/williamlsh/vault

ENV GO111MODULE=on

COPY go.mod .

COPY go.sum .

RUN go mod download

FROM build_base AS server_builder

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -a -tags vaultd -ldflags '-w -extldflags "-static"' ./cmd/vaultd

FROM alpine as vaultd

COPY --from=server_builder /go/bin/vaultd /bin/vaultd

ENTRYPOINT [ "/bin/vaultd" ]