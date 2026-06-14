# Subagent-Driven Development Execution Plan

> 基于 `refactor-complete-redesign` 变更提案，使用子代理驱动开发框架执行 8 阶段 57 项任务。

## 调度总览

| 调度# | 任务 | 名称 | 模型等级 | 预估子代理次数 |
|-------|------|------|---------|-------------|
| S1 | 1.1-1.5 | Phase 1 基础框架 | 最强模型 | impl×1 + review×2 |
| S2 | 2.1-2.3 | Storage 核心 (repo/index/lock) | 最强模型 | impl×1 + review×2 |
| S3 | 2.4-2.5 | Parser + Version | 标准模型 | impl×1 + review×2 |
| S4 | 3.1 | Source 接口 + 工厂 | 最强模型 | impl×1 + review×2 |
| S5 | 3.2-3.4 | 三个 Resolver 实现 | 标准模型 | impl×1 + review×2 |
| S6 | 3.5 | Validator | 标准模型 | impl×1 + review×2 |
| S7 | 4.1-4.2 | Installer + Symlink | 最强模型 | impl×1 + review×2 |
| S8 | 4.3-4.5 | Copy + Sync + Agent | 标准模型 | impl×1 + review×2 |
| S9 | 5.1-5.3 | Operations 三文件 | 标准模型 | impl×1 + review×2 |
| S10 | 6.1-6.4 | CLI 基础 + 4 命令 | 最强模型 | impl×1 + review×2 |
| S11 | 6.5-6.10 | CLI 中间 6 命令 | 标准模型 | impl×1 + review×2 |
| S12 | 6.11-6.14 | CLI 末尾 4 命令 | 标准模型 | impl×1 + review×2 |
| S13 | 7.1-7.2 | 前端脚手架 + 布局 | 最强模型 | impl×1 + review×2 |
| S14 | 7.3-7.5 | 前端 3 页面 | 标准模型 | impl×1 + review×2 |
| S15 | 7.6-7.7 | 前端 2 页面 | 标准模型 | impl×1 + review×2 |
| S16 | 7.8-7.11 | 组件 + Hooks + Bridge + API | 最强模型 | impl×1 + review×2 |
| S17 | 8.1-8.4 | 单元测试 | 标准模型 | impl×1 + review×2 |
| S18 | 8.5-8.7 | 集成/E2E/打包 | 最强模型 | impl×1 + review×2 |

**总计**: 18 个子代理调度 × 每次 3 个子代理(impl + spec-review + quality-review) = ~54 次子代理调用

## 依赖关系

```
S1 ──► S2 ──► S3 ──► S4 ──► S5 ──► S6
        │                     │
        ▼                     ▼
S7 ──► S8 ──► S9 ──► S10 ──► S11 ──► S12
                              │
                              ▼
                    S13 ──► S14 ──► S15 ──► S16
                                          │
                                          ▼
                                    S17 ──► S18
```

- S1 是全局依赖，必须在所有调度的前面
- S2/S3 依赖 S1 的 models.go
- S4 依赖 S1 的 models.go
- S5/S6 依赖 S4 的 Resolver 接口
- S7/S8 依赖 S2/S3 的 Storage + S4 的 Source
- S9 相对独立
- S10-S12 依赖 S2-S9 全部后端域
- S13-S16 前端独立于后端（仅通过 bridge.ts 调用 API）
- S17-S18 依赖全部实现

## 调度详细设计

### S1: Phase 1 基础框架（最强模型）

**任务**: 1.1, 1.2, 1.3, 1.4, 1.5
**工作目录**: `backend/`
**上下文**: 完全没有现有代码，全新初始化的项目骨架

**上下文传递给子代理**:
- 设计文档: `docs/superpowers/specs/2026-06-11-skills-manager-redesign.md`
- models.go 的完整类型定义
- config.go 的接口签名
- repository.go 的接口签名
- go.mod module path

**输出产物**:
- `backend/go.mod` → 初始化 Go module
- `backend/cmd/skill/main.go` → 空 cobra 根命令
- `backend/pkg/models/models.go` → 全量数据模型
- `backend/internal/operations/config.go` → 配置读写
- `backend/internal/storage/repository.go` → 仓库初始化

### S2: Storage 核心（最强模型）

**任务**: 2.1, 2.2, 2.3
**工作目录**: `backend/internal/storage/`
**依赖**: S1 产物

### S3: Parser + Version（标准模型）

**任务**: 2.4, 2.5
**工作目录**: `backend/internal/storage/`
**依赖**: S1 产物（models.go）

### S4: Source 接口（最强模型）

**任务**: 3.1
**工作目录**: `backend/internal/source/`
**依赖**: S1 产物（models.go）

### S5: Resolver 实现（标准模型）

**任务**: 3.2, 3.3, 3.4
**工作目录**: `backend/internal/source/`
**依赖**: S4 产物（Resolver 接口）

### S6: Validator（标准模型）

**任务**: 3.5
**工作目录**: `backend/internal/source/`
**依赖**: S1 产物（parser.go）

### S7: Installer + Symlink（最强模型）

**任务**: 4.1, 4.2
**工作目录**: `backend/internal/distribute/`
**依赖**: S2, S4

### S8: Copy + Sync + Agent（标准模型）

**任务**: 4.3, 4.4, 4.5
**工作目录**: `backend/internal/distribute/`
**依赖**: S7

### S9: Operations（标准模型）

**任务**: 5.1, 5.2, 5.3
**工作目录**: `backend/internal/operations/`
**依赖**: S1, S2

### S10: CLI 基础 + 核心命令（最强模型）

**任务**: 6.1, 6.2, 6.3, 6.4
**工作目录**: `backend/cmd/skill/`
**依赖**: S2, S4, S9

### S11: CLI 中间命令（标准模型）

**任务**: 6.5, 6.6, 6.7, 6.8, 6.9, 6.10
**工作目录**: `backend/cmd/skill/`
**依赖**: S10

### S12: CLI 末尾命令（标准模型）

**任务**: 6.11, 6.12, 6.13, 6.14
**工作目录**: `backend/cmd/skill/`
**依赖**: S10, S11

### S13: 前端脚手架（最强模型）

**任务**: 7.1, 7.2
**工作目录**: `frontend/`
**依赖**: S1 (无代码依赖，并行度高)

### S14: 前端页面 1（标准模型）

**任务**: 7.3, 7.4, 7.5
**工作目录**: `frontend/src/pages/`
**依赖**: S13

### S15: 前端页面 2（标准模型）

**任务**: 7.6, 7.7
**工作目录**: `frontend/src/pages/`
**依赖**: S13

### S16: 前端组件/Hooks/Bridge/API（最强模型）

**任务**: 7.8, 7.9, 7.10, 7.11
**工作目录**: `frontend/src/` + `backend/pkg/api/`
**依赖**: S14, S15

### S17: 单元测试（标准模型）

**任务**: 8.1, 8.2, 8.3, 8.4
**工作目录**: `backend/`
**依赖**: S1-S12

### S18: 集成/E2E/打包（最强模型）

**任务**: 8.5, 8.6, 8.7
**工作目录**: `backend/` + `frontend/`
**依赖**: S16, S17

## 质量门禁

每个调度完成后，必须经历两阶段审查：
1. **Spec 合规审查**: 对照 spec 文件逐条核对，不信任实现者的报告
2. **代码质量审查**: 检查命名、结构、测试完整性

只有两个审查都通过后，才能进入下一个调度。

## 执行优先级

- S1-S4: **最高优先级** — 所有后端域的基础
- S5-S9: **高优先级** — 填补后端各领域
- S10-S12: **中优先级** — CLI 接口层
- S13-S16: **中优先级** — 前端可视化（可与 S10-S12 并行？不，子代理不能并行）
- S17-S18: **低优先级** — 收尾阶段