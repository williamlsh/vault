FROM golang

ADD . /go/src/github.com/williamlsh/vault

RUN go install github.com/williamlsh/vault/cmd/vaultd

ENV VAULTD_LOG_LEVEL=all

EXPOSE 8080-8082

ENTRYPOINT [ "/go/bin/vaultd" ]
