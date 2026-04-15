.PHONY: migrate-up migrate-down migrate-status

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/travel_planner?sslmode=disable

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status
