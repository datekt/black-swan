BINARY_CLI=bse-cli
BINARY_WEB=bse-web

.PHONY: all build run-cli run-web test fmt vet clean help

all: help

build:
	@echo "==> Compiling BSE CLI..."
	go build -v -o $(BINARY_CLI) ./apps/cli
	@echo "==> Compiling BSE Webview Server..."
	go build -v -o $(BINARY_WEB) ./apps/webview
	@echo "==> Build complete!"

run-cli:
	go run ./apps/cli --preset presets/solar_system.json

run-web:
	go run ./apps/webview

test:
	go test -v -race ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	@go clean
	@-rm -f $(BINARY_CLI)
	@-rm -f $(BINARY_WEB)
	@-rm -f $(BINARY_CLI).exe
	@-rm -f $(BINARY_WEB).exe
	@echo "==> Cleaned build artifacts"

help:
	@echo Description: Black Swan Event (BSE) Automation Tasks
	@echo.
	@echo Usage: make [target]
	@echo.
	@echo Targets:
	@echo   all       Show this help
	@echo   build     Build CLI and Webview binaries
	@echo   run-cli   Run CLI with the solar_system preset
	@echo   run-web   Run the local webview server
	@echo   test      Run all unit tests with race detector
	@echo   fmt       Format code with gofmt
	@echo   vet       Run go vet static analysis
	@echo   clean     Remove build artifacts