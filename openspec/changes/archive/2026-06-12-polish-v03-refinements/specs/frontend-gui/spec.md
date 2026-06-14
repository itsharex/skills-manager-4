## ADDED Requirements

### Requirement: Agent Grouping
The system SHALL group agents sharing the same skills directory path.

#### Scenario: Same path agents grouped
- **WHEN** two or more agents share the same `SkillsDir`
- **THEN** they display as a single row in the agent list
- **AND** the display name SHALL be the shortest agent name among the group
- **AND** hovering SHALL show a tooltip with all agent names joined by ", "

#### Scenario: No detected agents in group
- **WHEN** none of the agents in a group are detected on the filesystem
- **THEN** the display name SHALL be the shortest name followed by " 等"
- **AND** the tooltip SHALL still show all agent names

#### Scenario: Detected groups sorted first
- **WHEN** displaying agent groups
- **THEN** groups with at least one detected agent SHALL appear before undetected groups

### Requirement: Agent Selection Persistence
The system SHALL remember the user's last agent selection for skill installation.

#### Scenario: Last selection restored
- **WHEN** user opens the install page after a previous installation
- **THEN** the agents that were selected last time SHALL be pre-checked
- **AND** the selection SHALL be stored in localStorage with key `skillsmanager:lastAgentSelection`

#### Scenario: Clear and reselect
- **WHEN** user clicks "清除选择"
- **THEN** all agent checkboxes SHALL be unchecked
- **AND** the localStorage entry SHALL be cleared

### Requirement: Skill Pool Global Scan
The system SHALL scan all known agent global skills directories when no project path is specified.

#### Scenario: Global scan shows pool and new skills
- **WHEN** user scans with empty project path
- **THEN** skills already in the index SHALL be displayed under "已收录"
- **AND** skills found in agent directories but not in index SHALL be displayed under "新发现"

### Requirement: Skill Pool Project Scan
The system SHALL scan agent subdirectory names within a project path for undiscovered skills.

#### Scenario: Project scan finds new skills
- **WHEN** user specifies a project path
- **THEN** the system SHALL iterate through all known agent subdirectory names (e.g., `.claude/skills`, `.cursor/skills`)
- **AND** only return skills NOT already in the index
- **AND** already-indexed skills SHALL NOT be shown

## MODIFIED Requirements

### Requirement: Agent List Display
The system SHALL detect and display known AI coding agents.

#### Scenario: Agent list shows 40+ agents
- **WHEN** user views the agent list
- **THEN** at least 40 known agents SHALL be listed
- **AND** agents with detected skills directories SHALL be visually marked (green dot)

### Requirement: Skill Installation
The system SHALL install skills from a specified source and distribute to selected agents.

#### Scenario: Install with agent selection
- **WHEN** user searches and finds skills in MarketPage
- **THEN** an agent selection panel SHALL appear before the install button
- **AND** user SHALL select which agent groups to install to
- **AND** the selected agent IDs SHALL be passed to the backend install call