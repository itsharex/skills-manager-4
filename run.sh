#!/bin/bash

# Skills Manager - 构建和运行脚本

echo "正在构建和运行 Skills Manager..."

# 检查前端是否已构建
if [ ! -d "frontend/dist" ]; then
    echo "正在构建前端..."
    cd frontend
    npm run build
    cd ..
fi

# 运行应用 (production 模式，嵌入前端)
echo "正在启动应用..."
CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
CGO_CFLAGS="-Wno-deprecated-declarations" \
go run -tags production .