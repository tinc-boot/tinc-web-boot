install: backend

tools:
	GO111MODULE=off go get -u -v github.com/go-bindata/go-bindata/...
	GO111MODULE=off go get -u -v github.com/reddec/struct-view/cmd/events-gen
	GO111MODULE=off go get -u -v github.com/reddec/jsonrpc2/cmd/...

ui:
	cd web/ui && npm i && npm run build

regen: tools ui
	PATH="$(PATH):$(GOPATH)/bin" go generate web/routes.go
	PATH="$(PATH):$(GOPATH)/bin" go generate web/internal/*.go
	PATH="$(PATH):$(GOPATH)/bin" go generate network/event_types.go

backend: regen
	go build -o tinc-web-boot -v ./cmd/tinc-web-boot/main.go

.PHONY: install