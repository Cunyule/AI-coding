.PHONY: test build run compose

test:
	go test ./...

build:
	go build -o bin/fingerprint-server ./cmd/server
	go build -o bin/fingerprint-client ./cmd/client

run:
	go run ./cmd/server

compose:
	docker compose up --build
