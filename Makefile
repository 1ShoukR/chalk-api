# Makefile for local development

.PHONY: run
run:
	@echo "🚀 Running API locally..."
	go run .

.PHONY: dev
dev:
	@echo "🚀 Running API with hot reload (requires air)..."
	air

.PHONY: docker-up
docker-up:
	@echo "🐳 Starting Docker Compose..."
	docker compose up -d --build

.PHONY: docker-down
docker-down:
	@echo "🛑 Stopping Docker Compose..."
	docker compose down

.PHONY: docker-logs
docker-logs:
	@echo "📜 Tailing Docker Compose logs..."
	docker compose logs -f

.PHONY: docker-rebuild
docker-rebuild:
	@echo "🔄 Rebuilding and restarting containers..."
	docker compose down && docker compose up -d --build

.PHONY: psql
psql:
	@echo "🐘 Connecting to PostgreSQL..."
	docker exec -it chalk-api-postgres-1 psql -U postgres -d chalkdb

.PHONY: test
test:
	@echo "🧪 Running tests..."
	go test ./...

.PHONY: build
build:
	@echo "🔨 Building..."
	go build -o bin/chalk-api .

.PHONY: clean
clean:
	@echo "🧹 Cleaning..."
	rm -rf bin/ tmp/
