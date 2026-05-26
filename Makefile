.PHONY: migrate-up migrate-down migrate-status

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/travel_planner?sslmode=disable

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status

.PHONY: mocks test test-cover cover-business cover-full cover-html clean generate

UNAME_S := $(shell uname -s)
ifeq ($(OS),Windows_NT)
    MOCKS_SCRIPT := scripts/generate_mocks.ps1
    MOCKS_CMD := powershell -ExecutionPolicy Bypass -File $(MOCKS_SCRIPT)
else
    MOCKS_SCRIPT := scripts/generate_mocks.sh
    MOCKS_CMD := bash $(MOCKS_SCRIPT)
endif

mocks:
	@echo "Generating mocks using $(MOCKS_SCRIPT)..."
	@chmod +x $(MOCKS_SCRIPT) 2>/dev/null || true
	@$(MOCKS_CMD)
	@echo "Mocks generated."


generate:
	@echo "Installing easyjson..."
	@go install github.com/mailru/easyjson/easyjson@latest
	@echo "Running go generate..."
	@go generate ./internal/dto/...
	@echo "easyjson generation complete."

test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

cover-business:
	go test ./... \
	  -coverpkg=guidely-app/internal/album,guidely-app/internal/album/repository,guidely-app/internal/review,guidely-app/internal/review/repository,guidely-app/internal/auth,guidely-app/internal/auth/repository,guidely-app/internal/service,guidely-app/internal/repository,guidely-app/internal/handlers,guidely-app/internal/middleware,guidely-app/pkg/config,guidely-app/pkg/db,guidely-app/pkg/utils \
	  -coverprofile=coverage.out \
	  && go tool cover -func=coverage.out | grep total
