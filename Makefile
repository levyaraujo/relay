include .env

MIGRATIONS_DIR=database/migrations

.PHONY: migrate-up migrate-down migrate-create migrate-force db-reset run test

migrate-up:
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) up

migrate-down:
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) down 1

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) $(name)

migrate-force:
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) force $(version)

db-reset:
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) drop -f
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) up

run:
	go run .

test:
	go test ./...
