#!/bin/bash
set -e

if ! command -v mockgen &> /dev/null; then
    echo "mockgen not found. Install with: go install github.com/golang/mock/mockgen@latest"
    exit 1
fi

# Auth service repositories
mkdir -p internal/auth/repository/mocks
mockgen -destination internal/auth/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/auth/repository UserRepository,SessionRepository

# Album service repositories
mkdir -p internal/album/repository/mocks
mockgen -destination internal/album/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/album/repository AlbumRepository

# Gateway repositories (Place, Trip, Review, Category, User, Session)
mkdir -p internal/repository/mocks
mockgen -destination internal/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/repository PlaceRepository,TripRepository,ReviewRepository,CategoryRepository,UserRepository,SessionRepository

# Gateway services (Place, Profile, Trip, Review, Category)
mkdir -p internal/service/mocks
mockgen -destination internal/service/mocks/mock_service.go \
    -package mocks \
    guidely-app/internal/service PlaceService,ProfileService,TripService,ReviewService,CategoryService

echo "Mocks generated successfully!"