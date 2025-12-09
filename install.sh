#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' 

info() {
    echo -e "${GREEN}ℹ️  $1${NC}"
}

warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $OS in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        msys*|mingw*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Неподдерживаемая ОС: $OS"
            exit 1
            ;;
    esac

    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        arm*)
            ARCH="arm"
            ;;
        *)
            error "Неподдерживаемая архитектура: $ARCH"
            exit 1
            ;;
    esac

    PLATFORM="${OS}/${ARCH}"
    BINARY_NAME="git-checkpoint"
    ARCHIVE_NAME="git-checkpoint-${OS}-${ARCH}"
    [ "$OS" = "windows" ] && ARCHIVE_NAME="git-checkpoint.exe" || ARCHIVE_NAME="${ARCHIVE_NAME}.tar.gz"
}

get_latest_release() {
    info "Получение информации о последнем релизе..."

    RELEASE_INFO=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest")

    if [ $? -ne 0 ]; then
        error "Не удалось получить информацию о релизе"
        exit 1
    fi

    VERSION=$(echo "$RELEASE_INFO" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    DOWNLOAD_URL=$(echo "$RELEASE_INFO" | grep "browser_download_url.*${ARCHIVE_NAME}" | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$VERSION" ] || [ -z "$DOWNLOAD_URL" ]; then
        error "Не удалось найти подходящий бинарник для платформы $PLATFORM"
        exit 1
    fi

    info "Найдена версия: $VERSION"
}

install_binary() {
    info "Скачивание $ARCHIVE_NAME..."

    TMP_DIR=$(mktemp -d)
    ARCHIVE_PATH="$TMP_DIR/$ARCHIVE_NAME"

    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$ARCHIVE_PATH" "$DOWNLOAD_URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$ARCHIVE_PATH" "$DOWNLOAD_URL"
    else
        error "Не найден curl или wget для скачивания"
        exit 1
    fi

    info "Распаковка..."
    cd "$TMP_DIR"
    if [ "$OS" = "windows" ]; then
        cp "$ARCHIVE_PATH" "$TMP_DIR/git-checkpoint.exe"
        EXTRACTED_BINARY="git-checkpoint.exe"
    else
        tar -xzf "$ARCHIVE_NAME"
        EXTRACTED_BINARY="git-checkpoint-${OS}-${ARCH}"
    fi

    if [ -w "/usr/local/bin" ] || [ -w "/usr/local" ]; then
        INSTALL_DIR="/usr/local/bin"
        SUDO=""
    else
        INSTALL_DIR="$HOME/bin"
        mkdir -p "$INSTALL_DIR"
        
        export PATH="$INSTALL_DIR:$PATH"
        
        SHELL_RC=""
        USER_SHELL=$(basename "$SHELL")
        case "$USER_SHELL" in
            zsh)
                SHELL_RC="$HOME/.zshrc"
                ;;
            bash)
                SHELL_RC="$HOME/.bashrc"
                ;;
            *)
                SHELL_RC="$HOME/.profile"
                ;;
        esac
        
        if [ -f "$SHELL_RC" ] && ! grep -q "$INSTALL_DIR" "$SHELL_RC"; then
            echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_RC"
            info "Добавлено в PATH в файл $SHELL_RC"
        fi
        
        warn "Установка в $INSTALL_DIR. Перезапустите терминал или выполните: source $SHELL_RC"
    fi

    info "Установка в $INSTALL_DIR..."
    if [ -n "$SUDO" ]; then
        sudo cp "$EXTRACTED_BINARY" "$INSTALL_DIR/git-checkpoint"
        sudo chmod +x "$INSTALL_DIR/git-checkpoint"
    else
        cp "$EXTRACTED_BINARY" "$INSTALL_DIR/git-checkpoint"
        chmod +x "$INSTALL_DIR/git-checkpoint"
    fi

    cd /
    rm -rf "$TMP_DIR"
}


main() {
    GITHUB_REPO="safevabecodehub/git-checkpoint"

    info "Установка Git Checkpoint TUI..."

    detect_platform
    info "Обнаружена платформа: $PLATFORM"

    get_latest_release
    install_binary

    info "🎉 Установка завершена! Запустите 'git-checkpoint' для начала работы."
}

main "$@"