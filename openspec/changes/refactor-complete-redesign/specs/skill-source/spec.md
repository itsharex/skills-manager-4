## ADDED Requirements

### Requirement: Multi-Source Skill Discovery
The system SHALL discover skills from multiple sources via a unified Resolver interface.

#### Scenario: GitHub repository resolver
- **WHEN** a GitHub URL (e.g. `github.com/owner/repo`) is provided
- **THEN** the system SHALL detect whether it is a single-skill or multi-skill repository
- **AND** shallow clone (`--depth 1`) to a temporary directory
- **AND** return one or more `ResolvedSkill` objects

#### Scenario: HTTP registry resolver
- **WHEN** an HTTP registry URL (e.g. `https://skills.sh`) is provided
- **THEN** the system SHALL fetch the registry index
- **AND** resolve individual skill entries

#### Scenario: Local path resolver
- **WHEN** a local file system path is provided
- **THEN** the system SHALL detect whether it is a directory, a ZIP file, or a SKILL.md file
- **AND** handle each case appropriately

#### Scenario: Unsupported source
- **WHEN** an unsupported source URL is provided
- **THEN** the system SHALL return an error indicating no resolver matched

### Requirement: Resolver Interface
The system SHALL define a `Resolver` interface with `Resolve` and `CanHandle` methods.

#### Scenario: Resolver factory
- **WHEN** `NewResolver(source)` is called
- **THEN** the system SHALL iterate registered resolvers
- **AND** return the first resolver whose `CanHandle` returns true
- **OR** return an error if no resolver matches

### Requirement: SKILL.md Validation
The system SHALL validate that resolved skills contain a properly formatted SKILL.md file.

#### Scenario: Valid SKILL.md
- **WHEN** a resolved skill has a SKILL.md with valid frontmatter (name, description, version)
- **THEN** the system SHALL pass validation

#### Scenario: Invalid SKILL.md
- **WHEN** a resolved skill has missing or malformed frontmatter
- **THEN** the system SHALL return a validation error with details