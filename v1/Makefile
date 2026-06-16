all: build

build:
	@echo "Building the binaries..."
	@go build -o main ./dev/cmd/server

run:
	@go run ./dev/cmd/server

docker-run:
	@echo "Running docker compose..."
	@docker compose up 2>/dev/null

docker-down:
	@echo "Shutting down docker compose"
	@docker compose down 2>/dev/null