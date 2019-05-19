SHELL := /bin/bash

.EXPORT_ALL_VARIABLES:
SRC_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
OUT_DIR := $(SRC_DIR)/_output
BIN_DIR := $(OUT_DIR)/bin
GOFLAGS := -mod=vendor
GO111MODULE := on

$(@info $(shell mkdir -p $(OUT_DIR) $(BIN_DIR)))

.PHONY: build
build:
	go build -o $(BIN_DIR)/xp .

.PHONE: test
test:
	go test -covermode=count -coverprofile=$(OUT_DIR)/coverage.out .

.PHONY: coveralls
coveralls:
	goveralls -coverprofile=$(OUT_DIR)/coverage.out -service=travis-ci -repotoken $COVERALLS_TOKEN
