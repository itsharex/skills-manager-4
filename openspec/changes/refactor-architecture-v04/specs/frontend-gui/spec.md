## ADDED Requirements

### Requirement: Pool Path Configuration
The system SHALL allow the user to configure a local skill pool directory path in Settings.

#### Scenario: Configure pool path
- **WHEN** user navigates to Settings page
- **THEN** a "技能池路径" input SHALL be displayed
- **AND** the current pool path from config SHALL be pre-filled
- **AND** user SHALL be able to modify and save it

#### Scenario: Default pool path
- **WHEN** no pool path is configured
- **THEN** the system SHALL use `~/.skill-pool/` as the default
- **AND** the pool directory SHALL be created if it doesn't exist

### Requirement: Market Source Management
The system SHALL allow the user to configure multiple market sources in Settings.

#### Scenario: Add market source
- **WHEN** user clicks "添加来源" in Settings
- **THEN** a form SHALL appear with Name, URL, Type fields
- **AND** Type SHALL be one of: "pool", "github", "registry"
- **AND** the source SHALL be saved to config

#### Scenario: Manage market sources
- **WHEN** viewing Settings
- **THEN** all configured market sources SHALL be listed
- **AND** each source SHALL have enable/disable toggle and remove button

### Requirement: Skill Pool Display
The system SHALL display all skills found in the configured pool path.

#### Scenario: Show pool contents
- **WHEN** user navigates to SkillsPoolPage
- **THEN** the page SHALL read the configured PoolPath from config
- **AND** display all skill directories (containing SKILL.md) under that path

#### Scenario: Empty pool state
- **WHEN** the pool path contains no skill directories
- **THEN** a friendly empty state SHALL be displayed with instructions to import skills

### Requirement: Local Machine Scan
The system SHALL scan all agent global skill directories and optionally project directories, matching results against the pool.

#### Scenario: Scan all agents
- **WHEN** user clicks "本机扫描" on SkillsPoolPage
- **THEN** all KnownAgents' SkillsDir directories SHALL be scanned
- **AND** results SHALL be compared against pool → categorized as "已收录" or "未收录"
- **AND** results SHALL be displayed in two columns

#### Scenario: Scan with project path
- **WHEN** user specifies a project path during scan
- **THEN** agent skill subdirectory names SHALL be matched under the project path
- **AND** combined with global agent scan results
- **AND** matched against pool → categorized

#### Scenario: Import from scan
- **WHEN** user clicks "导入到池" on an 未收录 skill from scan results
- **THEN** the skill directory SHALL be copied/symlinked into PoolPath
- **AND** the pool display SHALL refresh

### Requirement: Market Search with Priority
The system SHALL search enabled market sources with local pool priority.

#### Scenario: Search all sources
- **WHEN** user clicks "搜索所有" on MarketPage
- **THEN** enabled sources SHALL be searched in order: pool → github → registry
- **AND** results SHALL be merged with source name labels

#### Scenario: Search single source
- **WHEN** user clicks search on a specific source
- **THEN** only that source SHALL be searched
- **AND** results displayed with source name

### Requirement: Config Save API
The system SHALL expose a SaveConfig API from the backend.

#### Scenario: Save config
- **WHEN** user modifies Settings and clicks save
- **THEN** the frontend SHALL call `saveConfig(cfg)` on the backend
- **AND** the backend SHALL persist the config to disk

## MODIFIED Requirements

### Requirement: Agent Detection
The system SHALL detect AI coding agents by checking for the agent binary or application directory, not by checking the skills directory.

#### Scenario: CLI agent detection
- **WHEN** detecting a CLI-based agent (e.g., codex, claude code)
- **THEN** the system SHALL check if the binary is in PATH via `exec.LookPath`
- **AND** the agent SkillsDir SHALL be returned regardless of detection result

#### Scenario: IDE agent detection
- **WHEN** detecting an IDE/desktop agent (e.g., Cursor, Claude Desktop)
- **THEN** the system SHALL check if the application directory exists via `os.Stat`

#### Scenario: Agent list display
- **WHEN** user views the agent list
- **THEN** detected agents SHALL have green indicator (exe/application found)
- **AND** not-detected agents SHALL have gray indicator (binary not installed)
- **AND** SkillsDir SHALL be shown for all agents

### Requirement: Skill Pool Page
The system SHALL provide a skill pool page that displays the configured local pool directory contents and allows local machine scanning.

#### Scenario: Pool path from config
- **WHEN** SkillsPoolPage loads
- **THEN** it SHALL fetch the configured PoolPath via `getConfig()`
- **AND** display skills found in that directory
- **AND** provide a "本机扫描" section to scan agent/project dirs

## REMOVED Requirements

### Requirement: ScanPool with Dynamic Path Input
**Reason**: 拆分为两部分 — PoolPath 从配置读取（池页面），扫描独立为 ScanLocal
**Migration**: SkillsPoolPage 不再接受手动输入的扫描路径，改为使用配置的 PoolPath + 独立的"本机扫描"操作