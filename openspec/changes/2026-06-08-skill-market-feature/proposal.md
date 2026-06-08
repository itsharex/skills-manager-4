## Why

当前 Skills Manager 只支持单个技能的安装和管理，缺少以下核心功能：
- 技能市场的发现和分类浏览
- 技能市场 URL 的灵活配置
- Agent 全局技能库的扫描和展示
- 项目技能的迁移和软链管理
- 更全面的 Agent 支持（OpenClaw、Hermes、Antigravity 等）
- 批量技能同步功能

这些功能将显著提升 Skills Manager 的可用性和覆盖范围。

## What Changes

- **新增技能市场功能**：扫描技能市场并分类展示
- **技能市场配置**：支持灵活配置技能市场 URL
- **升级已安装技能展示**：扫描 Agent 全局技能库并合并去重，展示 Agent 安装状态
- **新增项目扫描功能**：扫描项目技能，支持物理迁移到技能库并保留软链
- **扩展 Agent 列表**：全面支持市场主流 Agent（OpenClaw、Hermes、Antigravity、Codex、Opencode 等）
- **批量同步功能**：支持选择多个技能批量同步到选定 Agent

## Capabilities

### New Capabilities
- `skill-market`: 技能市场的浏览、搜索和分类展示
- `market-config`: 技能市场 URL 的配置管理
- `global-skill-scan`: Agent 全局技能库扫描与合并去重
- `project-skill-import`: 项目技能扫描、迁移与软链管理
- `agent-extended`: 全面的主流 Agent 支持
- `batch-sync`: 多技能批量同步功能

### Modified Capabilities

## Impact

- 后端 API：新增多个 API 接口（市场扫描、项目扫描、批量同步等）
- 前端 UI：新增市场页面、项目扫描页面、升级已安装技能展示页面
- 数据模型：扩展配置模型支持技能市场、项目技能等
