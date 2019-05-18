SHELL := /bin/bash

.EXPORT_ALL_VARIABLES:
SRC_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
OUT_DIR := $(SRC_DIR)/_output
BIN_DIR := $(OUT_DIR)/bin

$(@info $(shell mkdir -p $(OUT_DIR) $(BIN_DIR)))

.PHONY: build
build:
	go build -mod=vendor -o $(BIN_DIR)/xp .

.PHONE: test
test:
	go test -mod=vendor .
