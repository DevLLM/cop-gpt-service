# Set default args for build
build_args ?=-o ./build/

.PHONY: help
help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  dev       to start development server"
	@echo "  get-copilot-token       to get Github Copilot Plugin Token"
	@echo "  build     to build binary. Use build_args to set build args. Default is '-o ./build/'"

.PHONY: dev
dev:
	@echo "Starting development server..."
	@go run main.go

.PHONY: get-copilot-token
get-copilot-token:
	@echo "Getting Github Copilot Plugin Token..."
	@bash ./shells/get_copilot_token.sh

.PHONY: build
build:
	@echo "Building binary..."
	@go build ${build_args} ./
