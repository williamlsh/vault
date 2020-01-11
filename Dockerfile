FROM golang:1.13-alpine

RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    && update-ca-certificates

RUN adduser -D -g '' vaultd

WORKDIR /go/src/github.com/williamlsh/vault/

COPY go.mod .

RUN go mod download
RUN go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/vaultd ./cmd/vaultd

FROM scratch

COPY --from=0 /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /etc/passwd /etc/passwd
COPY --from=0 /go/bin/vaultd /bin/vaultd

USER vaultd

ENTRYPOINT [ "/bin/vaultd" ]

EXPOSE 443 8080-8081

CMD ["-http-addr=:443", "-grpc-addr=:8080", "-prom-addr=:8081", "-tls-key=/testdata/server-key.pem", "-tls-cert=/testdata/server-cert.pem", "-pg-user=postgres", "-pg-password=postgres", "-pg-dbname=postgres", "-pg-host=localhost", "-pg-sslmode=disable", "-pg-port=5432"]
