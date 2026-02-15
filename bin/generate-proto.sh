#!/bin/bash
set -e

# Check for protoc
if ! command -v protoc &> /dev/null; then
  echo "Error: protoc is not installed. Run: brew install protobuf"
  exit 1
fi

# Install protoc Go plugins if not already installed
if ! command -v protoc-gen-go &> /dev/null; then
  echo "Installing protoc-gen-go..."
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
  echo "Installing protoc-gen-go-grpc..."
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Create output directory
mkdir -p gen/pb

# Generate Go code from proto files
protoc \
  --proto_path=proto \
  --go_out=gen/pb --go_opt=paths=source_relative \
  --go-grpc_out=gen/pb --go-grpc_opt=paths=source_relative \
  proto/*.proto

# Add Go dependencies
go get google.golang.org/protobuf
go get google.golang.org/grpc

echo "Proto generation complete. Output in gen/pb/"
