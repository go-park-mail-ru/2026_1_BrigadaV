#!/bin/bash
set -e

if ! command -v mockgen &> /dev/null; then
    echo "mockgen not found. Install with: go install github.com/golang/mock/mockgen@latest"
    exit 1
fi

mkdir -p internal/auth/repository/mocks
mockgen -destination internal/auth/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/auth/repository UserRepository,SessionRepository

mkdir -p internal/album/repository/mocks
mockgen -destination internal/album/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/album/repository AlbumRepository

mkdir -p internal/repository/mocks
mockgen -destination internal/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/repository PlaceRepository,TripRepository,ReviewRepository,CategoryRepository,UserRepository,SessionRepository

mkdir -p internal/service/mocks
mockgen -destination internal/service/mocks/mock_service.go \
    -package mocks \
    guidely-app/internal/service PlaceService,ProfileService,TripService,CategoryService

mkdir -p internal/review/repository/mocks
mockgen -destination internal/review/repository/mocks/mock_repository.go \
    -package mocks \
    guidely-app/internal/review/repository ReviewRepository

mockgen -destination pkg/pb/album/mock_album_client.go -package album guidely-app/pkg/pb/album AlbumServiceClient
mockgen -destination pkg/pb/review/mock_review_client.go -package review guidely-app/pkg/pb/review ReviewServiceClient

echo "Mocks generated successfully!"