tools:
	GO111MODULE=off go get -v github.com/go-bindata/go-bindata/...

dev-tools: tools
	GO111MODULE=off go get -v github.com/reddec/jsonrpc2/cmd/...

regen:
	go generate web/internal/*.go