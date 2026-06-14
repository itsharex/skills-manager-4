# Change: Refactor architecture - pool, market, agent detection

## Why

当前架构存在四层概念混淆问题：

1. **技能池**：当前被误当作"动态扫描操作"（ScanPool），实际应是一个**配置的本地目录**（类似 apt 本地缓存仓库）
2. **市场**：当前只有自由输入 URL 搜索，缺少配置化的多来源支持（本地池 → GitHub → 开放市场）
3. **扫描 vs 池**：本地扫描 agent 目录和池内容展示被混在一起
4. **Agent 检测**：仅检查 SkillsDir 目录存在，但大多数 CLI agent（codex, claude code 等）的 skills 目录不会自动创建，应改为检测 agent 二进制/应用本身是否安装

## What Changes

### Backend (models + waillib)
- **MODIFIED** `models.Config`: 新增 `PoolPath string` 字段
- **ADDED** `models.MarketSource` 市场来源类型
- **MODIFIED** `models.Config`: 新增 `MarketSources []MarketSource` 字段
- **ADDED** `App.GetConfig()` / `App.SaveConfig()` 双向配置读写
- **ADDED** `App.ScanLocal(projectPath string)` 独立扫描方法（与池无关）
- **MODIFIED** `App.ScanPool()` → 改为直接读取配置的 PoolPath 目录
- **MODIFIED** `distribute.DetectAgents()` → 改为检测二进制/应用安装
- **MODIFIED** `distribute.Agent`: 新增 `DetectCmd string`, `DetectPath string` 字段

### Frontend
- **MODIFIED** `SettingsPage`: 新增池路径配置 + 市场来源管理 UI
- **MODIFIED** `SkillsPoolPage`: 显示池目录内容 + 扫描匹配功能
- **MODIFIED** `MarketPage`: 使用配置的市场来源搜索（本地优先）
- **ADDED** `Storage` 相关 bridge API: `saveConfig`, `addMarketSource`, `removeMarketSource`
- **MODIFIED** `types.ts`: 新增 `MarketSource` 类型

## Impact

- Affected specs: `frontend-gui` (spec 需要完整重写)
- Affected code:
  - `backend/pkg/models/models.go` (Config 扩展)
  - `backend/internal/distribute/agent.go` (Agent 检测逻辑重写)
  - `backend/internal/operations/config.go` (保存配置)
  - `backend/pkg/waillib/app.go` (ScanPool 重写 + ScanLocal 新增)
  - `frontend/src/pages/SettingsPage.tsx` (配置管理)
  - `frontend/src/pages/SkillsPoolPage.tsx` (池展示 + 扫描)
  - `frontend/src/pages/MarketPage.tsx` (市场来源搜索)
  - `frontend/src/types.ts` (新增类型)
  - `frontend/src/bridge.ts` (新增 API)
- **BREAKING**: `models.Config` 结构变更，现有 `~/.skill-repo/config.json` 需要迁移