.PHONY: build install

build:
	go build -o lmux ./cmd/lmux

install:
	sh scripts/install.sh
