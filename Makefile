tools:
	GO111MODULES=off go get -v github.com/reddec/jsonrpc2/cmd/...


regen:
	go generate web/internal/*.go