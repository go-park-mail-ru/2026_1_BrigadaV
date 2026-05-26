#!/bin/bash
set -e

if ! command -v mockgen &> /dev/null; then
    echo "mockgen not found. Install with: go install github.com/golang/mock/mockgen@latest"
    exit 1
fi

# Auth service
mkdir -p internal/auth/repository/mocks
mockgen -destination internal/auth/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/auth/repository UserRepository,SessionRepository

# Album service
mkdir -p internal/album/repository/mocks
mockgen -destination internal/album/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/album/repository AlbumRepository

# Gateway repositories (добавлены TripMemberRepository и TripInviteRepository)
mkdir -p internal/repository/mocks
mockgen -destination internal/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/repository TripRepository,TripMemberRepository,TripInviteRepository,PlaceRepository,CategoryRepository,ReviewRepository,UserRepository,SessionRepository

# Gateway services (TripService теперь включает все методы шеринга)
mkdir -p internal/service/mocks
mockgen -destination internal/service/mocks/mock_service.go \
    -package mocks \
    guidely-app/internal/service PlaceService,ProfileService,TripService,ReviewService,CategoryService

# Review service
mkdir -p internal/review/repository/mocks
mockgen -destination internal/review/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/review/repository ReviewRepository

echo "Mocks generated successfully!"