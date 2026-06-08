# Version Management Specification

## Purpose
Provide complete version lifecycle management for installed skills.

---

## Requirements

### Requirement: List Skill Versions
System must display all installed versions of a given skill.

- **Version Enumeration**: List all version directories under `{skillspool}/{skill-name}/`.
- **Latest Indicator**: Visually indicate which version is currently marked as "latest".
- **Version Metadata**: Show installation date, source, size for each version.
- **Sort Order**: Display versions in descending order (newest first).

#### Scenario: List Versions for a Skill
- **WHEN** User selects a skill and clicks "Versions"
- **THEN** Display modal/page with version list
- **FOR EACH** version, show:
  - Version identifier (directory name)
  - Installation date
  - Source (GitHub URL or local path)
  - Size (bytes)
  - Whether it's the current "latest"
- **IF** only one version exists, still show the list (with "Only version" indicator)

---

### Requirement: Switch Skill Version
System must allow changing the "latest" version of a skill and updating all agent symlinks.

- **Update Latest Pointer**: Change `{skillspool}/{skill-name}/latest` to point to new version.
- **Update Agent Symlinks**: For each agent using this skill, update their symlink to point to new version.
- **Atomic Operation**: If agent symlink update fails, still update latest pointer but report partial failure.
- **No Version Deletion**: Switching version does NOT delete the old version directory.

#### Scenario: Switch to Different Version
- **WHEN** User selects a non-latest version and clicks "Set as Latest"
- **THEN** Show confirmation with old and new version numbers
- **WHEN** User confirms
- **THEN** Update `latest` file
- **FOR EACH** agent with this skill:
  - Update their symlink to point to new version
  - Track failures but continue with others
- **THEN** Show report: "Updated X agents, Y failed (list names)"

---

### Requirement: Delete Skill Version
System must allow删除 specific versions while preserving at least one version.

- **Protected Last Version**: Cannot delete the only remaining version of a skill.
- **Agent Awareness**: Deleting a version should warn if agents are currently using it.
- **Force Option**: Provide "Force Delete" that updates agents to use another version first.
- **Permanent Deletion**: No soft-delete or trash; files are removed immediately.

#### Scenario: Delete Non-Latest Version
- **WHEN** User selects a version (not latest) and clicks "Delete Version"
- **THEN** Show confirmation dialog with version and skill name
- **IF** agents are using this version:
  - Show warning: "X agents are using this version. They will be switched to [latest-other-version]."
  - Provide checkbox "Switch agents to alternative version"
- **WHEN** User confirms
- **THEN** If agents using this version and checkbox checked:
  - Switch agents to alternative version
- **THEN** Delete version directory
- **THEN** Show success/failure report

#### Scenario: Attempt to Delete Last Version
- **WHEN** User attempts to delete the only version of a skill
- **THEN** Show error: "Cannot delete the only version. Install an alternative version first."
- **AND** Disable delete button

---

### Requirement: Compare Skill Versions
System must provide a way to compare two versions of the same skill.

- **Side-by-Side**: Display metadata for two versions in adjacent columns.
- **Difference Highlighting**: Highlight fields that differ (e.g., different sources).
- **Quick Actions**: Allow switching latest or deleting from comparison view.

#### Scenario: Compare Two Versions
- **WHEN** User selects two versions and clicks "Compare"
- **THEN** Display comparison view:
  ```
  Version A (v1.2.3)    |    Version B (v1.2.4)
  Installed: 2024-01-15 |    Installed: 2024-02-20
  Source: github.com/x  |    Source: github.com/x
  Size: 2.3 MB         |    Size: 2.5 MB (+8%)
  Agents: 3            |    Agents: 3
  ```

---

### Requirement: Version Rollback
System must support rolling back a skill to a previous version.

- **One-Click Rollback**: User selects an older version and clicks "Rollback".
- **Updates All Agents**: All agents using this skill get switched to rollback version.
- **Creates Backup**: Before rollback, create backup copy of current latest.

#### Scenario: Rollback Skill Version
- **WHEN** User selects an older version and clicks "Rollback to This Version"
- **THEN** Show rollback confirmation with:
  - Current version
  - Target version
  - Number of affected agents
- **WHEN** User confirms
- **THEN** Create backup of current latest: `{skill-name}/backup-{timestamp}/
- **THEN** Switch latest pointer to target version
- **THEN** Update all agent symlinks
- **THEN** Show success report with backup location

---

### Requirement: Version Statistics
System must display aggregate statistics about skill versions.

- **Version Count**: Total versions per skill
- **Version Age**: Newest vs oldest version dates
- **Version Size Distribution**: Total size and average per version
- **Update Frequency**: Average time between version installations

#### Data: VersionStats Structure
```go
type VersionStats struct {
    SkillName      string
    VersionCount   int
    OldestVersion  string
    NewestVersion  string
    TotalSize      int64
    AverageSize    int64
    DaysBetweenUpdates float64
}
```
