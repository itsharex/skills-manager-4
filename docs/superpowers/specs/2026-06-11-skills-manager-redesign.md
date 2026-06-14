# Skills Manager 重构设计方案

> 基于 `reDesign.md` 的完整重构设计，对齐 AI Agent 技能安装工具的最终形态。

## 一、重构决策摘要

| 决策项 | 选择 | 说明 |
|--------|------|------|
| 重构范围 | **推倒重来** | 遵循 `reDesign.md` 完整设计，重新组织项目结构 |
| 技术栈 | **Wails + Go + React + TypeScript** | 延续已验证技术栈，复用 Wails 集成经验 |
| 模块划分 | **领域驱动设计（DDD）** | 技能来源 / 技能存储 / 技能分发 / 系统运维 |
| 存储目录 | `~/.skill-repo/` | 命名空间+版本化存储 |
| 安装模式 | 默认软链接 + `--copy` + 自动降级 | 对齐 reDesign 推荐方案 |
| 前端页面 | 5 页面（Dashboard / Market / Skills / Detail / Settings） | 覆盖完整操作流 |

## 二、整体架构

### 2.1 四领域分层

```
┌─────────────────────────────────────────────────────────────────┐
│                        交互层 (Presentation)                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │  CLI 命令    │  │ Wails 桌面端 │  │  HTTP API    │           │
│  │  (skill xxx) │  │ (React SPA)  │  │ (预留 MCP)   │           │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘           │
│         └─────────────────┼─────────────────┘                   │
│              ┌────────────▼────────────┐                        │
│              │     pkg/api (统一门面)   │                        │
│              └────────────┬────────────┘                        │
├───────────────────────────┼─────────────────────────────────────┤
│                核心服务层 (Core Domain)                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  ┌─────────────────┐ ┌──────────┐ ┌──────────┐           │   │
│  │  │   技能来源域     │ │  技能存储域│ │ 技能分发 │           │   │
│  │  │   source/        │ │  storage/ │ │ distribute│          │   │
│  │  │ ┌─────────────┐ │ │┌─────────┐│ │┌────────┐│           │   │
│  │  │ │ discovery   │ │ ││repository││ ││install ││           │   │
│  │  │ │ → github    │ │ ││→ index  ││ ││→ link  ││           │   │
│  │  │ │ → http      │ │ ││→ lock   ││ ││→ copy  ││           │   │
│  │  │ │ → local/zip │ │ ││→ parser ││ ││→ sync  ││           │   │
│  │  │ │ → validator │ │ ││→ version││ ││→ agent ││           │   │
│  │  │ └─────────────┘ │ │└─────────┘│ │└────────┘│           │   │
│  │  └─────────────────┘ └──────────┘ └──────────┘           │   │
│  │                                                           │   │
│  │  ┌─────────────────────────────────────────────────────┐  │   │
│  │  │              系统运维域 operations/                  │  │   │
│  │  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌──────┐         │  │   │
│  │  │  │ health │ │cleanup │ │ stats  │ │config│         │  │   │
│  │  │  └────────┘ └────────┘ └────────┘ └──────┘         │  │   │
│  │  └─────────────────────────────────────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                        外部资源层                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │ GitHub Repos │  │ HTTP Registry│  │ 本地文件/ZIP │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 领域职责

| 领域 | Package | 核心职责 |
|------|---------|---------|
| **技能来源** | `internal/source/` | 多源技能发现、获取、格式校验 |
| **技能存储** | `internal/storage/` | 本地仓库管理、索引、锁文件、版本化存储 |
| **技能分发** | `internal/distribute/` | 软链接/复制安装、多 Agent 同步、Agent 配置 |
| **系统运维** | `internal/operations/` | 健康检查、清理、统计、配置管理 |

## 三、技术选型

### 3.1 总体技术栈

| 层级 | 选型 | 版本 |
|------|------|------|
| 桌面框架 | Wails | v2.12.0 |
| 后端语言 | Go | 1.22+ |
| CLI 框架 | Cobra | v1.8.0 |
| 前端框架 | React | 18 |
| 类型系统 | TypeScript | 5.4+ |
| 构建工具 | Vite | 5 |
| UI 组件 | shadcn/ui + Tailwind CSS | latest |
| 代码编辑 | @monaco-editor/react | latest |
| YAML 解析 | gopkg.in/yaml.v3 | v3.0.1 |

### 3.2 新增依赖

| 库 | 用途 |
|----|------|
| `github.com/AlecAivazis/survey/v2` 或 `github.com/manifoldco/promptui` | CLI 交互式多选 |
| `github.com/mholt/archiver/v3` | ZIP 解压支持 |

## 四、项目目录结构

```
skillsmanager/
├── backend/                         # Go 后端
│   ├── cmd/
│   │   └── skill/                   # CLI 入口（二进制名: skill）
│   │       ├── main.go              # Cobra 命令注册
│   │       ├── cmd_install.go
│   │       ├── cmd_search.go
│   │       ├── cmd_list.go
│   │       ├── cmd_info.go
│   │       ├── cmd_sync.go
│   │       ├── cmd_config.go
│   │       ├── cmd_doctor.go
│   │       ├── cmd_export.go
│   │       ├── cmd_edit.go
│   │       └── cmd_init.go
│   ├── internal/
│   │   ├── source/                  # 【技能来源域】
│   │   │   ├── source.go            # Resolver 接口 + 工厂函数
│   │   │   ├── github.go            # GitHub 单/多技能仓库解析
│   │   │   ├── http.go              # HTTP Registry (skills.sh 等)
│   │   │   ├── local.go             # 本地文件/文件夹/ZIP 导入
│   │   │   └── validator.go         # SKILL.md 格式校验
│   │   ├── storage/                 # 【技能存储域】
│   │   │   ├── repository.go        # 本地仓库目录管理
│   │   │   ├── index.go             # index.json 全局索引 CRUD
│   │   │   ├── lock.go              # lock.json 安装锁定记录
│   │   │   ├── parser.go            # SKILL.md 解析
│   │   │   └── version.go           # 版本比较与排序
│   │   ├── distribute/              # 【技能分发域】
│   │   │   ├── installer.go         # 安装流程编排
│   │   │   ├── symlink.go           # 软链接安装 + Windows 降级
│   │   │   ├── copy.go              # 复制模式安装
│   │   │   ├── sync.go              # 多 Agent 批量同步
│   │   │   └── agent.go             # Agent 配置与自动检测
│   │   └── operations/              # 【系统运维域】
│   │       ├── health.go            # 健康检查 (doctor)
│   │       ├── cleanup.go           # 清理孤立软链/未使用技能
│   │       ├── stats.go             # 统计与仪表盘数据
│   │       └── config.go            # 配置读写管理
│   └── pkg/
│       ├── api/
│       │   └── api.go               # 统一 API 门面
│       └── models/
│           └── models.go            # 所有共享数据类型
├── frontend/                        # React + TypeScript + Vite
│   └── src/
│       ├── App.tsx                  # 路由 + 侧边栏导航
│       ├── pages/
│       │   ├── Dashboard.tsx        # 技能总览
│       │   ├── Market.tsx           # 技能市场
│       │   ├── Skills.tsx           # 已安装技能管理
│       │   ├── SkillDetail.tsx      # 技能详情 + 编辑
│       │   └── Settings.tsx         # Agent + 仓库配置
│       ├── components/
│       │   ├── layout/
│       │   │   ├── Sidebar.tsx
│       │   │   └── StatusBar.tsx
│       │   ├── skill/
│       │   │   ├── SkillCard.tsx
│       │   │   ├── SkillMetaForm.tsx
│       │   │   └── SkillVersionBadge.tsx
│       │   ├── agent/
│       │   │   ├── AgentSelector.tsx
│       │   │   └── AgentStatus.tsx
│       │   └── dialogs/
│       │       ├── InstallDialog.tsx
│       │       └── ConfirmDialog.tsx
│       ├── hooks/
│       │   ├── useSkills.ts
│       │   ├── useAgents.ts
│       │   └── useMarket.ts
│       ├── bridge.ts
│       ├── types.ts
│       └── index.css
├── main.go                          # Wails 入口
├── app.go                           # Wails App 绑定
├── embed.go                         # 前端产物嵌入
├── wails.json                       # Wails 配置
├── go.mod / go.sum
└── reDesign.md
```

## 五、数据模型

### 5.1 技能标识

```go
// 技能标识: namespace:name@version
// 例: github:anthropic/skills/pdf@1.0.0  /  local:my-skill@latest
type SkillID struct {
    Namespace string `json:"namespace"`
    Name      string `json:"name"`
    Version   string `json:"version"`
}
```

### 5.2 配置文件 (~/.skill-repo/config.json)

```go
type Config struct {
    RepoPath      string       `json:"repo_path"`       // 默认 ~/.skill-repo/
    InstallMode   string       `json:"install_mode"`    // "symlink" | "copy"
    AutoFallback  bool         `json:"auto_fallback"`   // 软链失败自动降级复制
    DefaultAgents []string     `json:"default_agents"`  // 默认目标 Agent 列表
    LinkTargets   []LinkTarget `json:"link_targets"`    // 注册的 Agent 目录
    Repositories  []RepoSource `json:"repositories"`    // 自定义技能仓库
    CacheTTL      int          `json:"cache_ttl"`       // 缓存过期时间（秒）
}

type LinkTarget struct {
    ID      string `json:"id"`
    Path    string `json:"path"`
    Enabled bool   `json:"enabled"`
}

type RepoSource struct {
    Name    string `json:"name"`
    URL     string `json:"url"`
    Type    string `json:"type"`     // "registry" | "github"
    Enabled bool   `json:"enabled"`
}
```

### 5.3 技能索引 (index.json)

```go
type Index struct {
    Version    int                  `json:"version"`
    LastUpdate string               `json:"last_update"`
    Skills     map[string]IndexEntry `json:"skills"`
}

type IndexEntry struct {
    Name          string   `json:"name"`
    Namespace     string   `json:"namespace"`
    Versions      []string `json:"versions"`
    Latest        string   `json:"latest"`
    Source        string   `json:"source"`
    SourceType    string   `json:"source_type"`
    InstalledSize string   `json:"installed_size"`
    Tags          []string `json:"tags"`
    Description   string   `json:"description"`
}
```

### 5.4 安装锁定 (lock.json)

```go
type LockFile struct {
    Version int                  `json:"version"`
    Skills  map[string]LockEntry `json:"skills"`
}

type LockEntry struct {
    SkillID     SkillID            `json:"skill_id"`
    InstalledAt string             `json:"installed_at"`
    Source      string             `json:"source"`
    Agents      []LockAgentBinding `json:"agents"`
}

type LockAgentBinding struct {
    AgentID string `json:"agent_id"`
    Path    string `json:"path"`
    Mode    string `json:"mode"` // "symlink" | "copy"
}
```

## 六、核心接口设计

### 6.1 来源域 — Resolver 接口

```go
// internal/source/source.go
type ResolvedSkill struct {
    LocalPath string             // 本地临时路径
    SkillInfo *models.SkillInfo  // 解析出的元数据
    Cleanup   func()             // 清理临时文件
}

type Resolver interface {
    Resolve(ctx context.Context, source string, opts ResolveOptions) ([]ResolvedSkill, error)
    CanHandle(source string) bool
}

type ResolveOptions struct {
    SubPath string
    Version string
    Ref     string
}

func NewResolver(source string) (Resolver, error)
```

### 6.2 存储域 — Repository 接口

```go
// internal/storage/repository.go
type Repository struct {
    Root string
}

// SkillPath: ~/.skill-repo/skills/{namespace}/{name}@{version}/
func (r *Repository) SkillPath(namespace, name, version string) string
func (r *Repository) Store(skill ResolvedSkill, namespace, version string) (string, error)
func (r *Repository) Remove(namespace, name, version string) error
func (r *Repository) UpdateLatest(namespace, name, version string) error
```

### 6.3 分发域 — Installer 接口

```go
// internal/distribute/installer.go
type Installer struct {
    repo      *storage.Repository
    index     *storage.Index
    lock      *storage.LockFile
    agents    *AgentManager
    resolvers []source.Resolver
}

func (i *Installer) Install(ctx context.Context, req InstallRequest) (*InstallResult, error)
func (i *Installer) Uninstall(ctx context.Context, req UninstallRequest) error
func (i *Installer) Sync(ctx context.Context, opts SyncOptions) (*SyncResult, error)
```

## 七、CLI 命令体系

```bash
# 仓库初始化与管理
skill init                              # 初始化 ~/.skill-repo/
skill config [get|set] <key> [value]    # 配置管理
skill repo [add|remove|list] <source>   # 管理自定义技能仓库

# 技能发现
skill search <query>                    # 搜索技能（多源并行）
skill list [--category <cat>]          # 列出已安装技能
skill info <skill>                     # 查看技能详情
skill show <skill>                     # 显示 SKILL.md 原始内容
skill validate <skill>                 # 验证 SKILL.md 格式

# 技能安装
skill install <source> [skill-name]    # 安装技能（交互式多选）
  --agents <list>, --scope, --project-dir, --copy, --version, --ref, --sub-path

# 技能生命周期
skill uninstall <skill> [--agents]     # 卸载技能
skill update [skill...]                # 更新技能
skill sync [--agents]                  # 批量同步

# 元数据编辑
skill edit <skill>                     # 交互式编辑 frontmatter
skill edit <skill> --description "..."
skill edit <skill> --editor

# 诊断与工具
skill doctor                           # 环境诊断
skill export [--format json|yaml]      # 导出技能列表
skill import <file>                    # 批量导入
skill stats [--skill <name>]           # 统计信息
```

## 八、前端设计

### 8.1 页面结构

| 页面 | 路由 | 功能 |
|------|------|------|
| Dashboard | `/` | 指标卡 + 技能分布图 + 活动时间线 |
| Market | `/market` | 多源搜索 + 分类浏览 + 快速安装 |
| Skills | `/skills` | 卡片/列表视图 + 筛选 + 批量操作 |
| SkillDetail | `/skills/:id` | 元数据展示 + Markdown 编辑预览 |
| Settings | `/settings` | Agent 管理 + 仓库配置 + 全局设置 |

### 8.2 核心组件

- **AgentSelector**: 多 Agent 多选组件（checkbox 列表）
- **SkillCard**: 技能卡片（支持列表/网格布局切换）
- **InstallDialog**: 安装弹窗（来源输入 + Agent 选择 + 选项配置）
- **SkillMetaForm**: frontmatter 编辑表单
- **SkillVersionBadge**: 版本标签

## 九、实施计划

| 阶段 | 名称 | 工期 | 关键产出 |
|------|------|------|----------|
| 1 | 基础框架 | 1.5 天 | 项目脚手架、models、config |
| 2 | 技能存储域 | 3 天 | repository、index、lock、parser、version |
| 3 | 技能来源域 | 3 天 | github、http、local、validator、resolver 工厂 |
| 4 | 技能分发域 | 3 天 | installer、symlink、copy、sync、agent |
| 5 | 系统运维域 | 2 天 | health、cleanup、stats |
| 6 | CLI 命令 | 3 天 | 16 个子命令 + 交互式选择 |
| 7 | 前端 GUI | 4 天 | 5 页面 + 组件 + Wails 绑定 |
| 8 | 测试与发布 | 3 天 | 单元/集成/E2E 测试 + 打包 |
| **合计** | | **22.5 天** | |

### 依赖关系

```
Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4 ──► Phase 5
                                                    │
                              ┌──────────────────────┘
                              ▼
                          Phase 6 ──► Phase 7 ──► Phase 8
```

### 质量门禁

| 门禁项 | 阈值 |
|--------|------|
| 单元测试通过率 | 100% |
| 代码覆盖率 | ≥ 80% |
| Lint 检查 | 0 错误 |
| 安全漏洞扫描 | 0 高危 |
| 跨平台兼容 | macOS/Linux/Windows |

## 十、软链接策略（对齐 reDesign 第四节）

- **默认模式**: 软链接，修改技能后所有 Agent 即时生效
- **强制复制**: `--copy` 参数用于共享技能库场景
- **自动降级**: Windows 或无权限环境自动降级为复制模式
- **最佳实践**: `ln -sfn` 覆盖创建、`skill doctor` 检查悬挂链接、建议只读技能使用软链接