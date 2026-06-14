# Change: polish-v03-refinements

## Why

v0.2 DDD 重构已完成，但存在以下问题影响使用体验：
1. 前端 mock/demo 数据在无 Wails 绑定时显示假数据，误导用户
2. Agent 列表覆盖不全，缺少近 10 个流行 AI 编码工具
3. Agent 展示缺少同路径合并逻辑，界面冗长且未安装的 agent 和已安装的混排
4. 技能池扫描逻辑不完整：无项目时不区分已收录/新发现，有项目时不走 agent 子目录
5. 安装技能流程缺少 agent 选择面板，用户无法指定目标 Agent

## What Changes

- 删除 `frontend/src/bridge.ts` 中所有 mock fallback 函数
- 扩展 `backend/internal/distribute/agent.go` `KnownAgents()` 至 40+ agent
- 前端 Agent 展示：同路径合并分组，已检测排前面，最短名显示 + tooltip
- 技能池扫描重构：无项目时扫描所有 agent 全局 SkillsDir，分"已收录"/"新发现"
- 有项目时沿 agent 子目录名扫描，只显示新发现技能
- MarketPage 安装前新增 Agent 选择面板，记住上次选择到 localStorage

## Impact

- Affected specs: `frontend-gui`, `skill-distribute`
- Affected code:
  - `frontend/src/bridge.ts` — 删除 mock
  - `frontend/src/types.ts` — 增强 AgentGroup
  - `frontend/src/pages/SettingsPage.tsx` — 增强分组展示
  - `frontend/src/pages/MarketPage.tsx` — 新增 agent 选择面板
  - `frontend/src/pages/SkillsPoolPage.tsx` — 区分全局/项目池
  - `frontend/src/components/AgentSelector.tsx` — 新建
  - `backend/internal/distribute/agent.go` — 扩展 agent 列表
  - `backend/pkg/waillib/app.go` — ScanPool 拆分为 scanGlobalPool/scanProjectPool