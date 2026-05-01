#!/bin/bash
set -e

PROTOC="/c/Users/egorp/protoc/bin/protoc.exe"

if [ ! -f "$PROTOC" ]; then
    echo "protoc not found at $PROTOC. Please install Protocol Buffers compiler."
    exit 1
fi

echo "Generating Go code from proto files..."
"$PROTOC" --go_out=. --go_opt=paths=source_relative \
          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
          proto/auth.proto proto/album.proto

echo "Proto generation complete."