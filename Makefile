.PHONY: migrate-up migrate-down migrate-status

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/travel_planner?sslmode=disable

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status


.PHONY: test test-cover test-race generate-mocks clean

generate-mocks:
	@echo "Generating mocks..."
	@mkdir -p internal/repository/mocks internal/service/mocks
	@mockgen -destination internal/repository/mocks/mock_repository.go -package mocks guidely-app/internal/repository UserRepository SessionRepository PlaceRepository TripRepository ReviewRepository
	@mockgen -destination internal/service/mocks/mock_service.go -package mocks guidely-app/internal/service AuthService PlaceService ProfileService TripService ReviewService
	@echo "Mocks generated successfully"

test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

test-race:
	go test ./... -race -v

clean:
	rm -f coverage.out coverage.html