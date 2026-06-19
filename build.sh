#!/bin/bash
#
# Skills Manager - 跨平台构建脚本
# 用法: ./build.sh [macos|linux|windows|all] [arm64|amd64]
#
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

BUILD_DIR="$SCRIPT_DIR/build/bin"
FRONTEND_DIR="$SCRIPT_DIR/frontend"
APP_NAME="skillsmanager"
VERSION="0.2.0"

# 颜色输出
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# ============================================================
# 1. 构建前端
# ============================================================
build_frontend() {
    if [ ! -d "$FRONTEND_DIR/dist" ] || [ "$FORCE_REBUILD" = "1" ]; then
        log_info "构建前端..."
        cd "$FRONTEND_DIR"
        npm install --silent
        npm run build
        cd "$SCRIPT_DIR"
    else
        log_info "前端已构建，跳过 (设置 FORCE_REBUILD=1 强制重构建)"
    fi
}

# ============================================================
# 2. 构建 macOS (arm64 / amd64)
# ============================================================
build_macos() {
    local arch="${1:-arm64}"
    local output="$BUILD_DIR/${APP_NAME}-darwin-${arch}"
    local app_bundle="$BUILD_DIR/Skills Manager.app"

    log_info "构建 macOS (${arch})..."
    mkdir -p "$BUILD_DIR"

    if [ "$arch" = "arm64" ]; then
        GOOS=darwin GOARCH=arm64 \
        CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
        CGO_CFLAGS="-Wno-deprecated-declarations" \
        go build -tags production -ldflags="-s -w" -o "$output" .
    elif [ "$arch" = "amd64" ]; then
        GOOS=darwin GOARCH=amd64 \
        CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
        CGO_CFLAGS="-Wno-deprecated-declarations" \
        go build -tags production -ldflags="-s -w" -o "$output" .
    else
        log_error "不支持的架构: $arch (支持 arm64/amd64)"
        exit 1
    fi

    log_info "macOS 可执行文件: $output"

    # 创建 .app Bundle
    log_info "打包 macOS .app Bundle..."
    rm -rf "$app_bundle"
    mkdir -p "$app_bundle/Contents/MacOS"
    mkdir -p "$app_bundle/Contents/Resources"

    cp "$output" "$app_bundle/Contents/MacOS/$APP_NAME"

    # Generate .icns from asklogo.png
    ICON_SRC="$SCRIPT_DIR/asklogo.png"
    if [ -f "$ICON_SRC" ]; then
        log_info "生成应用图标 (.icns)..."
        ICONSET_DIR="$BUILD_DIR/icon.iconset"
        rm -rf "$ICONSET_DIR"
        mkdir -p "$ICONSET_DIR"

        # Generate all required icon sizes using sips
        sips -z 16 16     "$ICON_SRC" --out "$ICONSET_DIR/icon_16x16.png"     >/dev/null 2>&1
        sips -z 32 32     "$ICON_SRC" --out "$ICONSET_DIR/icon_16x16@2x.png"  >/dev/null 2>&1
        sips -z 32 32     "$ICON_SRC" --out "$ICONSET_DIR/icon_32x32.png"     >/dev/null 2>&1
        sips -z 64 64     "$ICON_SRC" --out "$ICONSET_DIR/icon_32x32@2x.png"  >/dev/null 2>&1
        sips -z 128 128   "$ICON_SRC" --out "$ICONSET_DIR/icon_128x128.png"   >/dev/null 2>&1
        sips -z 256 256   "$ICON_SRC" --out "$ICONSET_DIR/icon_128x128@2x.png" >/dev/null 2>&1
        sips -z 256 256   "$ICON_SRC" --out "$ICONSET_DIR/icon_256x256.png"   >/dev/null 2>&1
        sips -z 512 512   "$ICON_SRC" --out "$ICONSET_DIR/icon_256x256@2x.png" >/dev/null 2>&1
        sips -z 512 512   "$ICON_SRC" --out "$ICONSET_DIR/icon_512x512.png"   >/dev/null 2>&1
        sips -z 1024 1024 "$ICON_SRC" --out "$ICONSET_DIR/icon_512x512@2x.png" >/dev/null 2>&1

        iconutil -c icns "$ICONSET_DIR" -o "$BUILD_DIR/iconfile.icns"
        cp "$BUILD_DIR/iconfile.icns" "$app_bundle/Contents/Resources/iconfile.icns"
        rm -rf "$ICONSET_DIR" "$BUILD_DIR/iconfile.icns"
        log_info "图标已打包到 .app Bundle"
    else
        log_warn "未找到图标文件: $ICON_SRC，跳过图标打包"
    fi

    cat > "$app_bundle/Contents/Info.plist" << PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIdentifier</key>
    <string>com.skillsmanager.app</string>
    <key>CFBundleName</key>
    <string>Skills Manager</string>
    <key>CFBundleDisplayName</key>
    <string>Skills Manager</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleIconFile</key>
    <string>iconfile</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSRequiresAquaSystemAppearance</key>
    <false/>
</dict>
</plist>
PLIST

    # Self-sign the app bundle (required for macOS to display icon and pass Gatekeeper)
    codesign --force --sign - "$app_bundle" 2>/dev/null && log_info "应用已自签名" || log_warn "自签名失败（不影响运行，但可能影响图标显示）"

    # Force macOS to recognize the new icon by clearing ALL relevant caches.
    # This is necessary because macOS LaunchServices aggressively caches icons by
    # bundle identifier, and simply rebuilding the .app at the same path does NOT
    # refresh the icon. This has been a recurring issue — we must clear caches
    # thoroughly every build to prevent "missing icon" reports.
    log_info "清理 macOS 图标缓存..."

    # 1. Unregister old entry (suppress errors if not registered yet)
    /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -u "$app_bundle" 2>/dev/null || true

    # 2. Re-register the new build
    /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -f "$app_bundle" 2>/dev/null && log_info "已注册到 LaunchServices" || log_warn "LaunchServices 注册失败"

    # 3. Clear the icon cache for this bundle identifier explicitly
    # The iconservices_agent caches icons separately from LaunchServices
    killall IconServicesAgent 2>/dev/null || true

    # 4. Restart Dock to pick up the new icon
    killall Dock 2>/dev/null && log_info "已刷新 Dock 图标缓存" || true

    # 5. Touch the app bundle to update modification date (helps cache invalidation)
    touch "$app_bundle"

    log_info "图标缓存清理完成（首次打开可能需要几秒加载图标）"

    # Remove raw binary - only keep the .app bundle
    rm -f "$output"

    log_info "macOS App Bundle: $app_bundle"
}

# ============================================================
# 3. 构建 Linux (amd64)
# ============================================================
build_linux() {
    local arch="${1:-amd64}"

    if [ "$(uname)" != "Linux" ]; then
        log_warn "当前不是 Linux 系统，需要交叉编译。"
        log_warn "Wails 依赖 CGO + WebKit2GTK，跨平台交叉编译较复杂。"
        log_warn "建议在 Linux 机器上原生编译，或使用 Docker："
        echo ""
        echo "  # Docker 方式构建 Linux 版本："
        echo "  docker run --rm -v \$(pwd):/app -w /app \\"
        echo "    -e GOOS=linux -e GOARCH=amd64 \\"
        echo "    golang:1.24-bookworm bash -c '"
        echo "      apt update && apt install -y libgtk-3-dev libwebkit2gtk-4.0-dev'"
        echo "      cd /app && go build -tags production -ldflags=\"-s -w\" -o build/bin/${APP_NAME}-linux-${arch} ."
        echo "    '"
        echo ""
        return 0
    fi

    local output="$BUILD_DIR/${APP_NAME}-linux-${arch}"
    log_info "构建 Linux (${arch})..."
    mkdir -p "$BUILD_DIR"

    GOOS=linux GOARCH=$arch \
    go build -tags production -ldflags="-s -w" -o "$output" .

    log_info "Linux 可执行文件: $output"
}

# ============================================================
# 4. 构建 Windows (amd64)
# ============================================================
build_windows() {
    local arch="${1:-amd64}"

    if [ "$(uname)" != "Linux" ] && [ "$(uname)" != "Darwin" ]; then
        # On Windows (MSYS2/Git Bash), cross-compilation is different
        log_info "构建 Windows (${arch})..."
        local output="$BUILD_DIR/${APP_NAME}-windows-${arch}.exe"
        mkdir -p "$BUILD_DIR"
        GOOS=windows GOARCH=$arch \
        go build -tags production -ldflags="-s -w -H windowsgui" -o "$output" .
        log_info "Windows 可执行文件: $output"
        return 0
    fi

    # Cross-compiling from macOS/Linux to Windows
    log_warn "从 $(uname) 交叉编译 Windows 需要 MinGW-w64 C 交叉编译器。"
    log_warn ""
    log_warn "  # macOS 安装交叉编译器："
    log_warn "  brew install mingw-w64"
    log_warn ""
    log_warn "  # 然后执行："
    log_warn "  CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \\"
    log_warn "    go build -tags production -ldflags=\"-s -w -H windowsgui\" \\"
    log_warn "    -o build/bin/${APP_NAME}-windows-${arch}.exe ."
    log_warn ""
    log_warn "  # 或者使用 Docker："
    log_warn "  docker run --rm -v \$(pwd):/app -w /app \\"
    log_warn "    -e CC=x86_64-w64-mingw32-gcc -e GOOS=windows -e GOARCH=amd64 \\"
    log_warn "    -e CGO_ENABLED=1 \\"
    log_warn "    x1unix/go-mingw:1.24 bash -c '"
    log_warn "      go build -tags production -ldflags=\"-s -w -H windowsgui\" -o build/bin/${APP_NAME}-windows-${arch}.exe ."
    log_warn "    '"
}

# ============================================================
# 主流程
# ============================================================
PLATFORM="${1:-macos}"
ARCH="${2:-arm64}"

echo "============================================"
echo "  Skills Manager v${VERSION} - 跨平台构建"
echo "============================================"
echo ""

# 构建前端
build_frontend

# 创建输出目录
mkdir -p "$BUILD_DIR"

case "$PLATFORM" in
    macos|darwin)
        build_macos "$ARCH"
        ;;
    linux)
        build_linux "$ARCH"
        ;;
    windows|win)
        build_windows "$ARCH"
        ;;
    all)
        log_info "构建全部平台..."
        build_macos arm64
        build_linux amd64
        build_windows amd64
        log_info "全部构建完成!"
        ls -lh "$BUILD_DIR/"
        ;;
    *)
        echo "用法: $0 [macos|linux|windows|all] [arm64|amd64]"
        echo ""
        echo "示例:"
        echo "  $0 macos              # 构建 macOS arm64"
        echo "  $0 macos amd64        # 构建 macOS Intel"
        echo "  $0 linux              # 构建 Linux amd64"
        echo "  $0 windows            # 构建 Windows amd64"
        echo "  $0 all                # 构建全部平台 (如果环境支持)"
        ;;
esac