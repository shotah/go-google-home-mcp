.PHONY: build test tidy lint run-mcp clean

BINARY ?= ghome
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build: ## Build ghome CLI/MCP binary
	go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/ghome

test: ## Run unit tests
	go test ./...

tidy: ## go mod tidy
	go mod tidy

run-mcp: build ## Run MCP server over stdio (requires prior ghome login)
	./$(BINARY) mcp

clean:
	rm -f $(BINARY) $(BINARY).exe
