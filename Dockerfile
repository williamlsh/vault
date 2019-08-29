FROM golang

WORKDIR $GOPATH/src/github.com/williamlsh/vault

COPY . .

RUN go get -d -v ./cmd/vaultd

RUN go install -v ./cmd/vaultd

ENTRYPOINT [ "go/bin/vaultd" ]