FROM golang

ADD . go/src/github.com/williamzion/vault

ENV VAULTD_LOG_LEVEL=all

# Install daemon vault command.
RUN go install github.com/williamzion/vault/cmd/vaultd

ENTRYPOINT /go/bin/vaultd

CMD [ "-http-addr", ":8080", "-grpc-addr", ":8081", "-key-file", "/go/src/vault/testdata/server-key.pem", "-cert-file", "/go/src/vault/testdata/server-cert.pem", "-pg-user", "postgres", "-pg-password", "password", "-pg-dbname", "postgres", "-pg-host", "", "-pg-sslmode", "", "-pg-port", "" ]
