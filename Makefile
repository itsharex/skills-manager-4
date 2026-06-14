# Skills Manager Makefile
# 用法: make <target>

APP_NAME    := skillsmanager
VERSION     := 0.2.0
BUILD_DIR   := build/bin
GO          := go
GO_TAGS     := production
LDFLAGS     := -s -w

# macOS 特定
MACOS_LDFLAGS := -framework UniformTypeIdentifiers
MACOS_CFLAGS  := -Wno-deprecated-declarations

.PHONY: all clean deps frontend build-macos build-linux build-windows run test

# ============================================================
# 默认: 构建 macOS arm64
# ============================================================
all: deps frontend build-macos

# ============================================================
# 依赖安装
# ============================================================
deps:
	@echo "安装 Go 依赖..."
	cd frontend && npm install --silent

# ============================================================
# 前端构建
# ============================================================
frontend:
	@echo "构建前端..."
	cd frontend && npm run build

# ============================================================
# 构建 macOS
# ============================================================
build-macos: frontend
	@echo "构建 macOS ($(shell uname -m))..."
	@mkdir -p $(BUILD_DIR)
	CGO_LDFLAGS="$(MACOS_LDFLAGS)" CGO_CFLAGS="$(MACOS_CFLAGS)" \
	$(GO) build -tags $(GO_TAGS) -ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(APP_NAME)-darwin-$(shell uname -m) .
	@echo "macOS 构建完成: $(BUILD_DIR)/"

# ============================================================
# 构建 macOS .app Bundle
# ============================================================
app-bundle: build-macos
	@echo "创建 macOS .app Bundle..."
	@rm -rf "$(BUILD_DIR)/Skills Manager.app"
	@mkdir -p "$(BUILD_DIR)/Skills Manager.app/Contents/MacOS"
	@mkdir -p "$(BUILD_DIR)/Skills Manager.app/Contents/Resources"
	@cp $(BUILD_DIR)/$(APP_NAME)-darwin-$(shell uname -m) "$(BUILD_DIR)/Skills Manager.app/Contents/MacOS/$(APP_NAME)"
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > "$(BUILD_DIR)/Skills Manager.app/Contents/Info.plist"
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> "$(BUILD_DIR)/Skills Manager.app/Contents/Info.plist"
	@echo '<plist version="1.0"><dict>' >> "$(BUILD_DIR)/Skills Manager.app/Contents/Info.plist"
	@echo '  <key>CFBundleExecutable</key><string>$(APP_NAME)</string>' >> "$(BUILD_DIR)/Skills Manager.app/Contents/Info.plist"
	@echo '  <key>CFBundleName</key><string>Skills Manager</string>' >> "$(BUILD_DIR)/Skills Manager.app/Contents/Info.plist"
	@echo '  <key>CFBundleVersion</key><string>$(VERSION)</string>' >> "$(BUILD_DIR)/Skills Manager.app/Contents/Info.plist"
	@echo '  <key>CFBundlePackageType</key><string>APPL</string>' >> "$(BUILD_DIR)/Skills Manager.app/Contents/Info.plist"
	@echo '  <key>LSMinimumSystemVersion</key><string>10.15</string>' >> "$(BUILD_DIR)/Skills Manager.app/Contents/Info.plist"
	@echo '  <key>NSHighResolutionCapable</key><true/>' >> "$(BUILD_DIR)/Skills Manager.app/Contents/Info.plist"
	@echo '</dict></plist>' >> "$(BUILD_DIR)/Skills Manager.app/Contents/Info.plist"
	@echo "App Bundle: $(BUILD_DIR)/Skills Manager.app"

# ============================================================
# 构建 Linux (需要 Linux 环境或 Docker)
# ============================================================
build-linux: frontend
	@echo "构建 Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -tags $(GO_TAGS) -ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .
	@echo "Linux 构建完成: $(BUILD_DIR)/"

# ============================================================
# 构建 Windows (需要 MinGW-w64 或 Docker)
# ============================================================
build-windows: frontend
	@echo "构建 Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build -tags $(GO_TAGS) -ldflags="$(LDFLAGS) -H windowsgui" \
		-o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .
	@echo "Windows 构建完成: $(BUILD_DIR)/"

# ============================================================
# 运行 (开发模式)
# ============================================================
run: frontend
	@echo "启动应用..."
	CGO_LDFLAGS="$(MACOS_LDFLAGS)" CGO_CFLAGS="$(MACOS_CFLAGS)" \
	$(GO) run -tags $(GO_TAGS) .

# ============================================================
# 测试
# ============================================================
test:
	@echo "运行测试..."
	$(GO) test ./backend/... -v -count=1

# ============================================================
# 清理
# ============================================================
clean:
	@echo "清理构建产物..."
	rm -rf $(BUILD_DIR)
	rm -rf frontend/dist
	@echo "清理完成"

# ============================================================
# Docker 构建 (Linux)
# ============================================================
docker-build-linux: frontend
	@echo "Docker 构建 Linux..."
	@mkdir -p $(BUILD_DIR)
	docker run --rm -v $(PWD):/app -w /app \
		-e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=1 \
		golang:1.24-bookworm bash -c '\
			apt-get update -qq && apt-get install -y -qq libgtk-3-dev libwebkit2gtk-4.0-dev && \
			go build -tags $(GO_TAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .'
	@echo "Linux Docker 构建完成: $(BUILD_DIR)/"

# ============================================================
# Docker 构建 (Windows)
# ============================================================
docker-build-windows: frontend
	@echo "Docker 构建 Windows..."
	@mkdir -p $(BUILD_DIR)
	docker run --rm -v $(PWD):/app -w /app \
		-e CC=x86_64-w64-mingw32-gcc -e GOOS=windows -e GOARCH=amd64 -e CGO_ENABLED=1 \
		x1unix/go-mingw:1.24 bash -c '\
			go build -tags $(GO_TAGS) -ldflags="$(LDFLAGS) -H windowsgui" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .'
	@echo "Windows Docker 构建完成: $(BUILD_DIR)/"

# ============================================================
# 帮助
# ============================================================
help:
	@echo "Skills Manager v$(VERSION) - 构建系统"
	@echo ""
	@echo "可用目标:"
	@echo "  make              构建 macOS arm64 (默认)"
	@echo "  make build-macos  构建 macOS (arm64 + amd64)"
	@echo "  make app-bundle   构建 macOS .app Bundle"
	@echo "  make build-linux  构建 Linux (需要 Linux 环境)"
	@echo "  make build-windows构建 Windows (需要 MinGW-w64)"
	@echo "  make run          运行应用 (macOS)"
	@echo "  make test         运行测试"
	@echo "  make clean        清理构建产物"
	@echo "  make docker-build-linux    Docker 构建 Linux"
	@echo "  make docker-build-windows  Docker 构建 Windows"
	@echo "  make help         显示此帮助"