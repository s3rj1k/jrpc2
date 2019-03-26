GO_BIN ?= go
CURL_BIN ?= curl
SHELL_BIN ?= sh

deps: check-gopath
	$(GO_BIN) get -u golang.org/x/net/context/ctxhttp

	$(CURL_BIN) -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | $(SHELL_BIN) -s -- -b ${GOPATH}/bin v1.15.0
	$(GO_BIN) get -u github.com/Quasilyte/go-consistent

test:
	$(GO_BIN) test -failfast ./...

lint:
	golangci-lint run
	go-consistent -pedantic -v ./...

check-gopath:
ifndef GOPATH
	$(error GOPATH is undefined)
endif