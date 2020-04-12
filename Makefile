tools:
	GO111MODULE=off go get -v github.com/go-bindata/go-bindata/...

dev-tools: tools
	GO111MODULE=off go get -u -v github.com/reddec/jsonrpc2/cmd/...
	GO111MODULE=off go get -u -v github.com/reddec/struct-view/cmd/events-gen

ui:
	cd web/ui && npm i && npm run build

regen:
	go generate web/internal/*.go
	go generate network/event_types.go