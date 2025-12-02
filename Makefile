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

test: ## 运行所有测试
	@go test -v ./tests/unit/... ./tests/integration/...

test-unit: ## 运行单元测试
	@go test -v ./tests/unit/... -cover

test-integration: ## 运行集成测试
	@go test -v ./tests/integration/...

test-e2e: ## 运行E2E测试（需要先启动前后端服务）
	@cd frontend && npm run test:e2e

test-coverage: ## 生成测试覆盖率报告
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

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

