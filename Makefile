# Development
.PHONY: dev
dev:
	go run cmd/api/main.go

# Build
.PHONY: build
build:
	go build -o bin/api cmd/api/main.go

# Frontend
.PHONY: frontend-install
frontend-install:
	cd frontend && npm install

.PHONY: frontend-dev
frontend-dev:
	cd frontend && npm run dev

.PHONY: frontend-build
frontend-build:
	cd frontend && npm run build

.PHONY: frontend-lint
frontend-lint:
	cd frontend && npm run lint

# Full stack development
.PHONY: install
install: frontend-install
	go mod download

.PHONY: build-all
build-all: frontend-build build

# Testing
.PHONY: test
test:
	go test -v -short ./...

.PHONY: test-unit
test-unit:
	go test -v -short $(shell go list ./internal/... | grep -v /test/)

.PHONY: test-integration
test-integration:
	go test -v -tags=integration ./tests/integration/... -timeout=10m

.PHONY: test-all
test-all: test-unit test-integration

.PHONY: test-coverage
test-coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic -short ./...
	go tool cover -html=coverage.out

.PHONY: test-integration-coverage
test-integration-coverage:
	go test -v -tags=integration -coverprofile=integration-coverage.out ./internal/test/... -timeout=10m
	go tool cover -html=integration-coverage.out -o integration-coverage.html

# Linting
.PHONY: lint
lint:
	golangci-lint run

# Clean
.PHONY: clean
clean:
	rm -rf bin/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/

# Docker
.PHONY: docker-build
docker-build:
	docker build -t tfa .

.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 fullstack-template

.PHONY: docker-dev
docker-dev:
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env; \
	fi
	docker-compose down --remove-orphans
	docker-compose up --build

.PHONY: docker-stop
docker-stop:
	docker-compose down

.PHONY: docker-logs
docker-logs:
	docker-compose logs -f

.PHONY: docker-clean
docker-clean:
	docker-compose down -v --rmi all --remove-orphans