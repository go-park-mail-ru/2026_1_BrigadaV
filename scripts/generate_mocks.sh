#!/bin/bash
set -e

if ! command -v mockgen &> /dev/null; then
    echo "mockgen not found. Install with: go install github.com/golang/mock/mockgen@latest"
    exit 1
fi

mkdir -p internal/repository/mocks
mkdir -p internal/service/mocks

mockgen -destination internal/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/repository UserRepository,SessionRepository,PlaceRepository,TripRepository,ReviewRepository

mockgen -destination internal/service/mocks/mock_service.go \
    -package mocks \
    guidely-app/internal/service AuthService,PlaceService,ProfileService,TripService,ReviewService

echo "Mocks generated successfully!"