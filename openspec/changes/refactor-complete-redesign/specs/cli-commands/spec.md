## ADDED Requirements

### Requirement: CLI Binary
The system SHALL provide a `skill` CLI binary built with Cobra.

#### Scenario: Root help
- **WHEN** `skill --help` is run
- **THEN** the system SHALL display available subcommands

### Requirement: Init Command
The system SHALL provide `skill init` to initialize the repository.

#### Scenario: Initialize repository
- **WHEN** `skill init` is run
- **THEN** the system SHALL create `~/.skill-repo/` directory structure
- **AND** create default config.json

### Requirement: Config Command
The system SHALL provide `skill config [get|set] <key> [value]` for configuration.

#### Scenario: Get config value
- **WHEN** `skill config get repo_path` is run
- **THEN** the system SHALL display the current value

#### Scenario: Set config value
- **WHEN** `skill config set install_mode copy` is run
- **THEN** the system SHALL update the configuration

### Requirement: Search Command
The system SHALL provide `skill search <query>` for multi-source skill discovery.

#### Scenario: Search skills
- **WHEN** `skill search pdf` is run
- **THEN** the system SHALL search across all configured sources
- **AND** display matching skills with source info

### Requirement: Install Command
The system SHALL provide `skill install <source> [skill-name]` with interactive multi-select.

#### Scenario: Install with interactive selection
- **WHEN** `skill install github.com/org/repo` is run
- **THEN** the system SHALL resolve the source
- **AND** if multiple skills found, present an interactive multi-select list
- **AND** install selected skills

#### Scenario: Install with flags
- **WHEN** `skill install github.com/org/repo --agents claude,cursor --copy` is run
- **THEN** the system SHALL use the specified agents and copy mode

### Requirement: Doctor Command
The system SHALL provide `skill doctor` for environment diagnostics.

#### Scenario: Run diagnostics
- **WHEN** `skill doctor` is run
- **THEN** the system SHALL check and report repository, symlink, and permission status

### Requirement: Export/Import Commands
The system SHALL provide `skill export` and `skill import` for batch operations.

#### Scenario: Export to JSON
- **WHEN** `skill export --format json` is run
- **THEN** the system SHALL output the installed skill list as JSON

#### Scenario: Import from file
- **WHEN** `skill import skills.yaml` is run
- **THEN** the system SHALL install all skills listed in the file

### Requirement: JSON Output Flag
All CLI commands SHALL support `--json` flag for machine-readable output.

#### Scenario: JSON output
- **WHEN** any command is run with `--json`
- **THEN** the system SHALL output structured JSON instead of formatted text