.DEFAULT_GOAL := help

.PHONY: help build install

help: ## Show available make commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

build: ## Build the lmux binary
	go build -o lmux ./cmd/lmux

install: ## Install lmux using scripts/install.sh
	sh scripts/install.sh
