## ADDED Requirements

### Requirement: Health Check (Doctor)
The system SHALL provide diagnostic checks for the skill environment.

#### Scenario: Full diagnostic
- **WHEN** `skill doctor` is run
- **THEN** the system SHALL check repository integrity, symlink validity, permission correctness, and index/lock consistency
- **AND** report any issues found

### Requirement: Cleanup
The system SHALL cleanup orphaned files and unused skills.

#### Scenario: Clean orphaned symlinks
- **WHEN** cleanup is triggered
- **THEN** the system SHALL remove symlinks that point to non-existent skill paths

#### Scenario: Clean unused skills
- **WHEN** cleanup is triggered with `--unused` flag
- **THEN** the system SHALL remove index entries referencing skills not installed to any agent
- **AND** optionally remove the repository directory

### Requirement: Statistics
The system SHALL collect and report usage statistics.

#### Scenario: Show stats
- **WHEN** `skill stats` is called
- **THEN** the system SHALL report total skill count, installed count, agent count, available updates, and storage usage

### Requirement: Configuration Management
The system SHALL manage user configuration at `~/.skill-repo/config.json`.

#### Scenario: Read config
- **WHEN** configuration is accessed
- **THEN** the system SHALL read from `~/.skill-repo/config.json`
- **AND** apply defaults for missing fields

#### Scenario: Write config
- **WHEN** a configuration value is changed
- **THEN** the system SHALL update `~/.skill-repo/config.json`