.PHONY: dev test seed docker-up docker-down

docker-up:
	docker compose up -d

docker-down:
	docker compose down

seed:
	cd backend && go run ./cmd/seed

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

dev: docker-up
	@echo "Mini HRM đang chạy trên Docker (DynamoDB: 8000, Admin: 8001, Cognito: 9229)"
