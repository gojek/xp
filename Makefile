.EXPORT_ALL_VARIABLES:
BIN_DIR := _bin

.PHONY: build
build:
	go build -mod=vendor -o $(BIN_DIR)/xp .

.PHONE: test
test:
	go test -mod=vendor .
