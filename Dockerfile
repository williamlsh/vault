FROM golang

WORKDIR /go/src/github.com/williamzion/vault

COPY . .

ENV VAULTD_LOG_LEVEL=all

# Install daemon vault command.
RUN go install ./cmd/vaultd

ENTRYPOINT /go/bin/vaultd

CMD [ "-http-addr", ":8080", "-grpc-addr", ":8081", "-key-file", "/go/src/vault/testdata/server-key.pem", "-cert-file", "/go/src/vault/testdata/server-cert.pem", "-pg-user", "postgres", "-pg-password", "password", "-pg-dbname", "postgres", "-pg-host", "database", "-pg-sslmode", "disable", "-pg-port", "5432" ]
