# Skill Market Specification

## Purpose
扩展 Skills Manager 的功能，提供技能市场发现、全局技能库扫描、项目技能管理、全面 Agent 生态支持和批量操作功能。

---

## Requirements

### Requirement: Skill Market Configuration and Discovery
技能市场 URL 必须可配置，并且能够从配置的 URL 扫描并分类展示技能。

- **Configurable Market URL**：系统必须提供配置界面让用户设置技能市场的 URL（支持本地目录或 GitHub 仓库）。
- **Category Discovery**：扫描过程必须识别技能的分类信息并按分类展示。
- **Skill Metadata Display**：每个技能必须显示完整的元数据（名称、描述、作者、版本、标签等）。

#### Scenario: Configure Skill Market URL
- **WHEN** 用户在设置页面配置技能市场 URL
- **THEN** 系统验证 URL 格式并保存到配置文件
- **THEN** 用户可以点击"刷新市场"按钮重新扫描

#### Scenario: Scan Skill Market and Categorize
- **WHEN** 用户点击"刷新市场"按钮
- **THEN** 系统从配置的 URL 位置扫描所有技能目录
- **THEN** 解析每个技能的 SKILL.md 获取元数据和分类
- **THEN** 在界面上按分类展示技能卡片

---

### Requirement: Global Skill Library Scan and Deduplication
系统必须扫描所有已配置 Agent 的全局技能库，合并去重并展示哪些 Agent 已安装。

- **Agent Global Path Scan**：系统必须按配置扫描每个 Agent 的全局技能库目录。
- **Deduplication**：相同名称的技能必须合并展示，不重复显示。
- **Agent Installation Status**：每个技能必须展示已安装的 Agent 列表。

#### Scenario: Scan All Agent Global Libraries
- **WHEN** 用户导航到"已安装技能"页面
- **THEN** 系统自动扫描所有 Agent 的全局技能库
- **THEN** 系统合并相同名称的技能
- **THEN** 展示技能列表，每个技能显示已安装的 Agent

#### Scenario: View Skill Agent Installation Details
- **WHEN** 用户点击某个技能
- **THEN** 显示该技能的详细信息和已安装 Agent 列表
- **THEN** 展示每个 Agent 的安装状态和路径

---

### Requirement: Project Skill Scanning and Migration
系统必须支持扫描指定项目目录中的技能，并提供选择是否将技能物理迁移到技能库，同时项目中保留软链。

- **Project Path Selection**：用户必须能选择要扫描的项目目录。
- **Skill Detection**：系统必须能扫描项目目录识别所有技能。
- **Migration Option**：对于每个技能，用户可以选择是否迁移到技能库。
- **Symlink Management**：迁移后项目中原位置必须保留软链指向技能库。

#### Scenario: Scan Project Directory
- **WHEN** 用户导航到"项目技能"页面并选择项目目录
- **THEN** 系统扫描该目录中的所有技能
- **THEN** 以列表形式展示项目中的技能

#### Scenario: Migrate Skill to Library with Symlink
- **WHEN** 用户选择项目中的技能并点击"迁移到技能库"
- **THEN** 系统将技能物理复制到技能库
- **THEN** 项目中原位置替换为指向技能库的软链
- **THEN** 确认迁移成功

---

### Requirement: Extended Agent Ecosystem Support with Search and Pagination
系统必须支持市面上所有主流 Agent 的完整列表，包括 Claude Code、GitHub Copilot、Cursor、Windsurf、Trae、Gemini CLI、OpenAI Codex、OpenClaw、Tabnine、Supermaven、Aider、CodeRabbit、Devin、Junie、Llion、Rogue、Hermes、Antigravity、SWE-agent、Opencode 等 Agent。

界面必须支持搜索和分页显示功能。

- **Comprehensive Agent List**：配置文件中必须预定义完整的主流 Agent 列表（20+ 个）。
- **Standard Directory Structure**：每个 Agent 必须有定义明确的技能库路径结构。
- **Auto-detection**：系统应能够自动检测哪些 Agent 已安装在本地。
- **Search Functionality**：界面必须支持通过 Agent 名称或 ID 进行快速搜索。
- **Pagination Display**：界面默认显示 10 个 Agent，点击"更多"按钮可瀑布流式显示更多（每次增加 10 个）。
- **Load More Button**：当还有更多 Agent 时显示"更多"按钮，否则不显示该按钮。

#### Scenario: Agent Configuration List with Search
- **WHEN** 用户查看 Agent 配置页面
- **THEN** 默认显示前 10 个 Agent
- **THEN** 显示搜索输入框支持快速搜索
- **WHEN** 用户输入搜索关键词
- **THEN** 实时过滤并显示匹配的 Agent（按名称或 ID 搜索）
- **THEN** 重置已显示的 Agent 数量

#### Scenario: Pagination with Load More
- **WHEN** 用户查看 Agent 配置页面且 Agent 数量超过 10 个
- **THEN** 显示"更多"按钮（显示剩余数量）
- **WHEN** 用户点击"更多"按钮
- **THEN** 额外显示 10 个 Agent
- **THEN** 更新剩余数量显示
- **WHEN** 所有 Agent 都已显示
- **THEN** 隐藏"更多"按钮

#### Scenario: Agent List Includes All Major Agents
- **WHEN** 用户查看 Agent 配置页面
- **THEN** 显示完整的 Agent 列表，包括：
  - Claude Code (Anthropic)
  - GitHub Copilot (Microsoft)
  - Cursor
  - Windsurf (Codeium)
  - Trae (字节跳动)
  - Gemini CLI (Google)
  - OpenAI Codex CLI
  - OpenClaw
  - Tabnine
  - Supermaven
  - Aider
  - CodeRabbit
  - Devin (Cognition)
  - Junie
  - Llion
  - Rogue
  - Hermes
  - Antigravity
  - SWE-agent
  - Opencode
- **THEN** 每个 Agent 显示检测状态和技能库路径

---

### Requirement: Batch Skill Synchronization
系统必须支持选择多个技能并批量同步到选定的 Agent。

- **Multi-selection UI**：界面必须支持选择多个技能。
- **Multi-agent Target**：用户必须能选择多个目标 Agent。
- **Progress Display**：批量同步过程中必须显示进度信息。

#### Scenario: Select and Sync Multiple Skills
- **WHEN** 用户在技能列表中选择多个技能并点击"批量同步"
- **THEN** 弹出 Agent 选择对话框
- **WHEN** 用户选择目标 Agent 并确认
- **THEN** 系统执行批量同步并显示进度
- **THEN** 完成后显示成功/失败统计
