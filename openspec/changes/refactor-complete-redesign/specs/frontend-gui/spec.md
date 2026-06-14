## ADDED Requirements

### Requirement: Dashboard Page
The system SHALL provide a dashboard page showing skill overview.

#### Scenario: Display metrics
- **WHEN** the dashboard page loads
- **THEN** the system SHALL display metric cards: total skills, agent count, available updates, storage usage

#### Scenario: Activity timeline
- **WHEN** the dashboard page loads
- **THEN** the system SHALL display a recent activity timeline

### Requirement: Market Page
The system SHALL provide a market page for browsing and discovering skills.

#### Scenario: Search and browse
- **WHEN** on the market page
- **THEN** the system SHALL support search input and source selection (GitHub/Registry/Local)
- **AND** display skill cards with an install button

### Requirement: Skills Page
The system SHALL provide a skills page for managing installed skills.

#### Scenario: List installed skills
- **WHEN** on the skills page
- **THEN** the system SHALL display installed skills in card or list view
- **AND** support filtering by agent, status, and tags
- **AND** support batch operations (uninstall, sync)

### Requirement: Skill Detail Page
The system SHALL provide a skill detail page with metadata and Markdown preview/edit.

#### Scenario: View skill detail
- **WHEN** viewing a skill's detail page
- **THEN** the system SHALL display metadata fields and rendered SKILL.md content

#### Scenario: Edit skill
- **WHEN** entering edit mode on the detail page
- **THEN** the system SHALL provide a form for frontmatter fields and a Monaco editor for Markdown

### Requirement: Settings Page
The system SHALL provide a settings page for agent and repository configuration.

#### Scenario: Manage agents
- **WHEN** on the settings page
- **THEN** the system SHALL display agent list with enable/disable toggles and path editing

#### Scenario: Manage repositories
- **WHEN** on the settings page
- **THEN** the system SHALL support adding/removing external skill repositories

### Requirement: Shared Components
The frontend SHALL use shared components for consistency.

#### Scenario: AgentSelector component
- **WHEN** an agent selection is needed
- **THEN** the system SHALL display a checkbox list of available agents

#### Scenario: SkillCard component
- **WHEN** displaying skill summaries
- **THEN** the system SHALL render skill card with name, description, version, and action buttons

### Requirement: Wails Binding
The frontend SHALL communicate with the Go backend through Wails bindings via a unified bridge.

#### Scenario: Call backend method
- **WHEN** a frontend action requires backend logic
- **THEN** the system SHALL call the Go method via `bridge.ts`
- **AND** handle the response or error