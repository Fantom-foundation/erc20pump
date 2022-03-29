# --------------------------------------------------------------------------
# Makefile for the ERC20 transfer transaction pump
#
# (c) Fantom Foundation, 2022
# --------------------------------------------------------------------------
# project related vars
PROJECT := $(shell basename "$(PWD)")

# go related vars
GO_BASE := $(shell pwd)
GO_BIN := $(CURDIR)/build

build/erc20pump:
	@go build -v \
	-o $@ \
	./cmd/erc20pump

.PHONY: build/erc20pump
all: help
help: Makefile
	@echo
	@echo "Choose a make command in "$(PROJECT)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
