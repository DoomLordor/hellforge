LOCAL_BIN := $(CURDIR)/bin

GOLANGCI_BIN := $(LOCAL_BIN)/golangci-lint

bin-deps:
	$(info #Installing project binary dependencies...)
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.1
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.0
	GOBIN=$(LOCAL_BIN) go install github.com/bufbuild/buf/cmd/buf@v1.61.0

test:
	$(info Running tests...)
	go test ./...

lint:
	$(info Running lint against all project files...)
	$(GOLANGCI_BIN) run --config=.golangci.yml ./...

generate-go:
	PATH="$(LOCAL_BIN):$$PATH" go generate ./internal/...

generate-proto:
	$(info Starting proto generation)
	PATH="$(LOCAL_BIN):$$PATH" buf generate --path ./api
