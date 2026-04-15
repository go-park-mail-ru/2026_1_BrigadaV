#!/bin/bash
set -e

if ! command -v mockgen &> /dev/null; then
    echo "mockgen not found. Install with: go install github.com/golang/mock/mockgen@latest"
    exit 1
fi

mkdir -p internal/repository/mocks
mkdir -p internal/service/mocks

echo "Generating repository mocks..."
mockgen -destination internal/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/repository UserRepository,SessionRepository,PlaceRepository,TripRepository,ReviewRepository

echo "Generating service mocks..."
mockgen -destination internal/service/mocks/mock_auth_service.go -package mocks guidely-app/internal/service AuthService
mockgen -destination internal/service/mocks/mock_place_service.go -package mocks guidely-app/internal/service PlaceService
mockgen -destination internal/service/mocks/mock_profile_service.go -package mocks guidely-app/internal/service ProfileService
mockgen -destination internal/service/mocks/mock_trip_service.go -package mocks guidely-app/internal/service TripService
mockgen -destination internal/service/mocks/mock_review_service.go -package mocks guidely-app/internal/service ReviewService

echo "Mocks generated successfully!"