## 1. Data Model and Configuration Extension
- [ ] 1.1 扩展 `Config` 数据模型，新增 `SkillMarket` 配置字段（URL、缓存设置等）
- [ ] 1.2 扩展 `Agent` 配置，新增完整的 Agent 列表（OpenClaw、Hermes、Antigravity、Codex、Opencode 等）
- [ ] 1.3 更新配置文件读取和保存逻辑

## 2. Skill Market Backend API
- [ ] 2.1 实现技能市场 URL 配置的 API（`GetMarketConfig`/`SetMarketConfig`）
- [ ] 2.2 实现技能市场扫描 API（`ScanMarket`），支持本地目录和 GitHub 仓库
- [ ] 2.3 实现技能市场分类展示 API（`ListMarketSkills`/`GetMarketCategories`）

## 3. Global Skill Library Scanning
- [ ] 3.1 实现 Agent 全局技能库扫描功能
- [ ] 3.2 实现技能去重逻辑
- [ ] 3.3 实现 Agent 安装状态查询 API（`ListGlobalSkillsWithAgents`）

## 4. Project Skill Import and Migration
- [ ] 4.1 实现项目目录扫描 API（`ScanProjectSkills`）
- [ ] 4.2 实现技能物理迁移功能（物理复制+软链）
- [ ] 4.3 实现跨平台软链支持（Windows 特殊处理）

## 5. Batch Synchronization
- [ ] 5.1 实现批量同步 API（`BatchSyncSkills`）
- [ ] 5.2 实现进度状态跟踪
- [ ] 5.3 实现同步结果统计和错误处理

## 6. Frontend - New Pages and Components
- [ ] 6.1 新增"技能市场"页面（URL 配置、分类展示、搜索等）
- [ ] 6.2 新增"项目技能"页面（目录选择、扫描展示、迁移功能）
- [ ] 6.3 升级"已安装技能"页面（全局扫描、Agent 状态展示）
- [ ] 6.4 新增批量选择和同步组件

## 7. Integration and Testing
- [ ] 7.1 完整功能集成测试
- [ ] 7.2 跨平台兼容性测试
- [ ] 7.3 CLI 命令补充
