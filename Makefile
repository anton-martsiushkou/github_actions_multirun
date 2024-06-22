export GOBIN := $(PWD)/bin
export PATH := $(GOBIN):$(PATH)
export GOFLAGS := -mod=mod

./bin:
	mkdir ./bin

./bin/multirun: ./bin
	go install ./cmd/multirun

./bin/golangci-lint: ./bin
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

./bin/minimock: ./bin
	go install github.com/gojuno/minimock/v3/cmd/minimock

./bin/gowrap: ./bin
	go install github.com/hexdigest/gowrap

.PHONY:
lint: ./bin/golangci-lint
	./bin/golangci-lint run --enable=goimports --disable=unused ./...

.PHONY:
test:
	 go test -v -race ./...

.PHONY:
generate: ./bin/minimock ./bin/gowrap
	go generate ./...

.PHONY:
all: generate lint test ./bin/multirun

.PHONY:
clean:
	rm -rf ./bin

