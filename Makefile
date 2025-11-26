.PHONY: help build run test migrate clean

help:
	@echo "Available commands:"
	@echo "  make build    - Build the application"
	@echo "  make run      - Run the server"
	@echo "  make test     - Run tests"
	@echo "  make migrate  - Run database migrations"
	@echo "  make clean    - Clean build artifacts"

build:
	@go build -o bin/server cmd/server/main.go
	@go build -o bin/migrate cmd/migrate/main.go

run:
	@go run cmd/server/main.go

test:
	@go test -v ./...

benchmark:
	@go test -bench=. -benchmem ./tests/...

migrate:
	@go run cmd/migrate/main.go

clean:
	@rm -rf bin/
	@rm -rf logs/

deps:
	@go mod download
	@go mod tidy

