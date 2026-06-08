# Skills Manager

跨 Agent 技能安装与管理工具。基于 Wails 构建，支持 GUI 桌面应用和 CLI 命令行两种使用方式。

## ✨ 功能特性

### 核心功能
- **多来源安装** - 支持从 GitHub 仓库、本地目录安装技能
- **跨 Agent 分发** - 一次安装，自动分发到 Trae、Claude Code、Cursor、Windsurf 等多个 Agent
- **版本管理** - 集中管理技能的多版本，支持版本切换
- **安装范围** - 支持全局安装（所有项目可用）或项目级安装（仅特定项目）
- **软链接分发** - 使用符号链接分发技能，节省存储空间

### 新增功能 (v0.2)
- **技能市场** - 配置技能市场 URL，发现并批量安装市场技能
- **项目扫描** - 扫描指定项目的现有技能，迁移到全局库
- **批量同步** - 选择多个技能批量同步到指定 Agent
- **扩展 Agent 列表** - 支持 20+ 主流 Agent（Claude Code、GitHub Copilot、Windsurf、Cursor、Trae、Gemini CLI、OpenAI Codex CLI、OpenClaw、Tabnine、Supermaven、Aider、CodeRabbit、Devin、Junie、Llion、Rogue、Hermes、Antigravity、SWE-agent、Opencode 等）

### 管理功能
- **清理工具** - 移除孤立软链接、清理未使用的技能
- **健康检查** - 检测和修复技能库问题
- **使用统计** - 查看技能使用情况的仪表盘
- **版本切换** - 在不同技能版本间切换

## 🛠️ 技术栈

- **后端**: Go 1.22, Wails v2
- **前端**: React 18, TypeScript, Vite
- **桌面框架**: Wails (Go + WebView2/WebKit)
- **设计风格**: 清新薄荷绿主题

## 📁 项目结构

```
skillsmanager/
├── backend/
│   ├── cmd/skills/          # CLI 命令行入口
│   │   └── main.go
│   ├── internal/
│   │   ├── agent/            # Agent 管理与技能分发
│   │   ├── config/           # 配置加载与保存
│   │   ├── installer/        # 技能安装核心逻辑
│   │   ├── lifecycle/        # 生命周期管理（清理、健康、统计）
│   │   ├── registry/        # 技能注册表管理
│   │   ├── skill/           # SKILL.md 解析
│   │   ├── storage/          # skillspool 存储
│   │   └── version/          # 版本比较
│   └── pkg/
│       ├── api/              # 统一 API 接口
│       └── models/           # 数据模型
├── frontend/
│   └── src/
│       ├── pages/
│       │   ├── InstallPage.tsx   # 安装技能页面
│       │   ├── SkillsPage.tsx    # 已安装技能列表
│       │   └── AgentsPage.tsx    # Agent 配置页面
│       ├── App.tsx
│       ├── bridge.ts         # Wails 桥接层
│       ├── types.ts
│       └── index.css         # 清新主题样式
├── openspec/                # OpenSpec 规范文档
│   └── changes/
│       ├── 2026-06-08-skill-market-feature/  # 技能市场功能
│       └── 2026-06-08-cleanup-and-more/      # 清理与管理功能
├── app.go                   # Wails 应用主入口
├── main.go                  # main 函数
├── embed.go                 # 前端资源嵌入
└── wails.json               # Wails 配置
```

## 🚀 安装与运行

### 前置条件

- Go 1.22+
- Node.js 18+
- Wails CLI

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

安装后，确保 `~/go/bin` 在 PATH 中：

```bash
# 验证安装
~/go/bin/wails version

# 或者添加到 PATH
export PATH="$HOME/go/bin:$PATH"
```

### 开发模式

```bash
# 安装前端依赖
cd frontend && npm install

# 返回项目根目录并启动开发服务器
cd .. && ~/go/bin/wails dev
```

### 构建应用

```bash
# 先构建前端
cd frontend && npm run build
cd ..

# 构建当前平台的应用
~/go/bin/wails build -s
```

构建产物位于 `build/bin/` 目录。

### 跨平台构建

Wails 支持同时构建多个平台的应用：

```bash
# 构建 macOS Intel 和 Apple Silicon 两个版本
~/go/bin/wails build -platform darwin/amd64,darwin/arm64 -s

# 构建 Windows 和 Linux 版本
~/go/bin/wails build -platform windows/amd64,linux/amd64 -s

# 构建所有主要平台
~/go/bin/wails build -platform darwin/amd64,darwin/arm64,windows/amd64,linux/amd64 -s
```

支持的平台：
- `darwin/amd64` - macOS Intel 64位
- `darwin/arm64` - macOS Apple Silicon (M1/M2/M3)
- `windows/amd64` - Windows Intel 64位
- `linux/amd64` - Linux Intel 64位
- `linux/arm64` - Linux ARM 64位

### 运行 GUI

方式1：直接运行已构建好的应用
```bash
open /Users/rain/sourceCode/ai/coding/skillsmanager/build/bin/Skills\ Manager.app
```

方式2：开发模式（推荐用于开发调试）
```bash
cd /Users/rain/sourceCode/ai/coding/skillsmanager
~/go/bin/wails dev
```

## 💻 CLI 使用

```bash
# 安装技能（GitHub 仓库）
skills install https://github.com/user/repo

# 安装技能（指定子目录）
skills install https://github.com/user/repo --sub-path skills/my-skill

# 安装技能（指定分支）
skills install https://github.com/user/repo --ref main

# 安装到特定 Agent
skills install https://github.com/user/repo --agent trae --agent claude

# 项目级安装
skills install https://github.com/user/repo --scope project --project-dir /path/to/project

# 列出已安装技能
skills list

# 列出 Agent 配置
skills agents

# 清理孤立软链接
skills cleanup

# 健康检查
skills health

# JSON 输出
skills list --json
skills agents --json
```

## 🖥️ GUI 使用

启动应用后，提供三个页面：

1. **安装技能** - 填写 GitHub URL 或本地路径，选择目标 Agent，设置安装范围
2. **已安装技能** - 查看所有已安装技能及其版本、分发状态
3. **Agent 配置** - 查看已检测到的 Agent 及技能目录配置（支持搜索、瀑布流加载）

### UI 特色
- 清新薄荷绿主题
- 平滑动画效果
- 响应式布局
- 搜索与筛选
- 卡片式视觉设计

## 📝 技能格式

技能目录需包含 `SKILL.md` 文件，格式如下：

```yaml
name: my-skill
description: 技能描述
version: 1.0.0
author: Author Name
tags:
  - frontend
  - ui
```

## ⚙️ 配置

配置文件位于 `~/.skillsmanager/config.yaml`（macOS）或对应系统路径：

```yaml
skillspool:
  root: ~/Library/Application Support/SkillsManager/skillspool

skill_market:
  url: https://skills.market.example.com  # 技能市场 URL

agents:
  trae:
    name: Trae
    skill_location: .trae/skills
    global_location: ~/.trae-cn/skills
  claude:
    name: Claude Code
    skill_location: .claude/skills
    global_location: ~/.claude/skills
  # ... 更多 Agent 配置
```

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 开发规范

- 后端遵循 Go 标准风格
- 前端使用 TypeScript 严格模式
- 新增功能建议先创建 OpenSpec 规范文档（参考 `openspec/changes/` 目录）

## 📄 License

MIT

---

**享受更高效的 Agent 技能管理！** 🎉
