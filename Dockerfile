FROM golang

ADD . go/src/github.com/williamzion/vault

# Install daemon vault command.
RUN go install github.com/williamzion/vault/vaultd

ENTRYPOINT [ "vaultd", "--http", "8080", "--grpc", "8081" ]

EXPOSE 8080 8081
