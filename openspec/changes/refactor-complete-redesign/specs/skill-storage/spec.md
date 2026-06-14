## ADDED Requirements

### Requirement: Versioned Repository Storage
The system SHALL store skills in a versioned, namespaced directory structure at `~/.skill-repo/skills/{namespace}/{name}@{version}/`.

#### Scenario: Store a skill
- **WHEN** a resolved skill is stored with namespace and version
- **THEN** the system SHALL copy the skill to `~/.skill-repo/skills/{namespace}/{name}@{version}/`
- **AND** update the latest symlink

#### Scenario: Remove a skill version
- **WHEN** a specific skill version is removed
- **THEN** the system SHALL delete the version directory
- **AND** update the index and lock entries

### Requirement: Global Index (index.json)
The system SHALL maintain a global index at `~/.skill-repo/index.json` tracking all stored skills.

#### Scenario: Add to index
- **WHEN** a skill is stored successfully
- **THEN** the system SHALL add or update the index entry with name, namespace, versions, latest, source, source_type, and description

#### Scenario: List all skills
- **WHEN** listing all skills from the index
- **THEN** the system SHALL return all index entries

### Requirement: Installation Lock (lock.json)
The system SHALL maintain a lock file at `~/.skill-repo/lock.json` tracking which skills are installed to which agents.

#### Scenario: Track installation
- **WHEN** a skill is installed to an agent
- **THEN** the system SHALL add a lock entry with skill_id, installed_at, source, and agent bindings

#### Scenario: Untrack uninstallation
- **WHEN** a skill is uninstalled from an agent
- **THEN** the system SHALL remove the agent binding from the lock entry
- **AND** remove the lock entry entirely if no agents remain

### Requirement: SKILL.md Parser
The system SHALL parse SKILL.md files to extract frontmatter metadata and markdown content.

#### Scenario: Parse frontmatter
- **WHEN** a SKILL.md file with YAML frontmatter is parsed
- **THEN** the system SHALL extract name, description, version, tags, author, and other metadata fields

### Requirement: Version Comparison
The system SHALL support semantic version comparison and sorting.

#### Scenario: Compare versions
- **WHEN** comparing versions "1.0.0" and "1.1.0"
- **THEN** the system SHALL correctly identify 1.1.0 as newer