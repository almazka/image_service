#!/bin/bash

echo "Generating protobuf files for image_service..."

# Добавляем GOPATH/bin в PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Проверяем установлен ли protoc
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed. Please install it first."
    echo "Ubuntu/Debian: sudo apt install protobuf-compiler"
    echo "macOS: brew install protobuf"
    exit 1
fi

# Проверяем установлены ли Go плагины
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Error: protoc-gen-go is not installed. Run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Error: protoc-gen-go-grpc is not installed. Run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

# Создаем директорию если не существует
mkdir -p proto

echo "Using protoc-gen-go: $(which protoc-gen-go)"
echo "Using protoc-gen-go-grpc: $(which protoc-gen-go-grpc)"

# Генерируем Go код
protoc --go_out=./proto --go_opt=paths=source_relative \
       --go-grpc_out=./proto --go-grpc_opt=paths=source_relative \
       proto/file_service.proto

if [ $? -eq 0 ]; then
    echo "Protobuf files generated successfully!"
    echo "Generated files:"
    ls -la proto/
else
    echo "Failed to generate protobuf files"
    exit 1
fi