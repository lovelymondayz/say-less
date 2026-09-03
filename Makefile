.PHONY: dev build up down logs clean deploy

# Start development environment
dev:
	cd backend && go run . &
	cd frontend && npm run dev
	@echo "Backend: http://localhost:8086 | Frontend: http://localhost:3006"

# Production build
build:
	cd frontend && npm ci && npm run build
	cd backend && go build -o sayless-api .
	@echo "Build complete"

# Docker operations
up:
	docker compose up -d --build
	@echo "Say Less running — FE: http://localhost:3006, BE: http://localhost:8086"

down:
	docker compose down

# Utility
logs:
	docker compose logs -f

clean:
	docker compose down -v
	rm -rf frontend/dist backend/sayless-api

deploy:
	./update.sh