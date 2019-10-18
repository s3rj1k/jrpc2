GO_BIN ?= go

export PATH := $(PATH):/usr/local/go/bin

all: lint

update:
	$(GO_BIN) get -u
	$(GO_BIN) get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GO_BIN) get -u github.com/mgechev/revive
	$(GO_BIN) mod tidy

test:
	$(GO_BIN) test -failfast ./...

lint:
	golangci-lint run ./...
	revive -config revive.toml ./...