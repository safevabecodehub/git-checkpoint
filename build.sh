#!/bin/bash

# Git Checkpoint TUI Build Script
# Кросс-компиляция для разных платформ

set -e

# Определение версий
VERSION=${1:-"v1.0.0"}
LDFLAGS="-s -w -X main.version=$VERSION"

# Целевые платформы
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm"
)

# Создание директории для релиза
mkdir -p dist

echo "🏗️  Сборка Git Checkpoint TUI версии $VERSION"

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    
    BINARY_NAME="git-checkpoint"
    if [ "$GOOS" = "windows" ]; then
        BINARY_NAME="git-checkpoint.exe"
    fi
    
    echo "📦 Сборка для $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="$LDFLAGS" -o "dist/$BINARY_NAME-$GOOS-$GOARCH" .
    
    # Создание архива
    cd dist
    if [ "$GOOS" = "windows" ]; then
        zip "$BINARY_NAME-$GOOS-$GOARCH.zip" "$BINARY_NAME-$GOOS-$GOARCH"
    else
        tar -czf "$BINARY_NAME-$GOOS-$GOARCH.tar.gz" "$BINARY_NAME-$GOOS-$GOARCH"
    fi
    cd ..
    
    echo "✅ $GOOS/$GOARCH готов"
done

echo ""
echo "🎉 Сборка завершена! Файлы в директории dist/"
echo "📋 Содержимое dist/:"
ls -la dist/