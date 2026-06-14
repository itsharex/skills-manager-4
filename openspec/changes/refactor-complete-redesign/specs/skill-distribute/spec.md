## ADDED Requirements

### Requirement: Install Orchestration
The system SHALL orchestrate the full install flow: resolve source, store in repository, install to agents.

#### Scenario: Install a skill
- **WHEN** `Install(ctx, req)` is called with source and agent targets
- **THEN** the system SHALL resolve the source via Resolver chain
- **AND** store the skill in the local repository
- **AND** install the skill to all specified agents
- **AND** update index.json and lock.json

### Requirement: Symlink Installation
The system SHALL install skills to agents via symlinks by default.

#### Scenario: Create symlink
- **WHEN** installing to an agent directory
- **THEN** the system SHALL create a symlink from `<agent-path>/<skill-name>` → `<repo-path>/<skill-path>`

#### Scenario: Windows fallback
- **WHEN** symlink creation fails (e.g. insufficient permissions or Windows without developer mode)
- **THEN** the system SHALL fall back to copy mode

### Requirement: Copy Installation
The system SHALL support copy-mode installation for shared skill repositories.

#### Scenario: Copy install
- **WHEN** `--copy` flag is set
- **THEN** the system SHALL copy skill files to the agent directory instead of symlinking

### Requirement: Multi-Agent Sync
The system SHALL synchronize installed skills to multiple agents.

#### Scenario: Sync all skills
- **WHEN** sync is triggered without specific skills
- **THEN** the system SHALL re-install all locked skills to their respective agents

#### Scenario: Sync specific skills
- **WHEN** sync is triggered with specific skill names
- **THEN** the system SHALL only re-install those skills

### Requirement: Agent Auto-Detection
The system SHALL detect installed AI agents and their configuration directories.

#### Scenario: Auto-detect agents
- **WHEN** initializing or running doctor
- **THEN** the system SHALL scan common agent directories (Claude, Cursor, etc.)
- **AND** populate config with discovered agents