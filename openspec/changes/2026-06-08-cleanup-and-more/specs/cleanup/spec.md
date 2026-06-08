# Skill Cleanup Specification

## Purpose
Provide safe, observable cleanup operations for skills across projects and global library.

---

## Requirements

### Requirement: Uninstall Skill from Project
System must allow removing a skill from a specific project's agent skills directory while preserving the global installation.

- **Project Scope Only**: Uninstalling from a project does NOT remove the skill from the global library.
- **Symlink Removal**: System must remove the symlink from `{project}/.agent/skills/` pointing to the skill.
- **Safe Operation**: Operation must fail gracefully if the symlink doesn't exist.
- **Confirmation**: UI must show confirmation dialog before uninstall.

#### Scenario: Uninstall Skill from Project
- **WHEN** User selects a skill in a project and clicks "Uninstall from Project"
- **THEN** Show confirmation dialog with skill name and project path
- **WHEN** User confirms
- **THEN** Remove symlink from project's skills directory
- **THEN** Show success message with count of removed symlinks
- **IF** symlink doesn't exist, show warning (not error)

---

### Requirement: Clean Orphaned Symlinks
System must detect and remove symlinks in agent skill directories that point to non-existent targets.

- **Automatic Detection**: System must identify symlinks where the target doesn't exist.
- **Safe Deletion**: Only broken symlinks are removed; valid symlinks and regular files are preserved.
- **Reporting**: Report the number of orphaned symlinks found and cleaned.
- **Agent Filtering**: Can target specific agent or all detected agents.

#### Scenario: Clean Orphaned Symlinks for Agent
- **WHEN** User navigates to Agent detail and clicks "Clean Orphaned Symlinks"
- **THEN** System scans the agent's skill directory for symlinks
- **THEN** For each symlink, verify the target exists
- **THEN** Remove broken symlinks
- **THEN** Display report: "Found X orphaned symlinks, cleaned Y"

#### Scenario: Clean All Agents
- **WHEN** User clicks "Clean All Agents" in bulk operations
- **THEN** System scans all detected agents' skill directories
- **THEN** Remove all broken symlinks
- **THEN** Display aggregated report per agent

---

### Requirement: Global Library Cleanup
System must identify and optionally remove skills that are not referenced by any agent.

- **Unused Skills Detection**: Find skills in global library with zero agent installations.
- **Orphaned Versions**: Find skill versions that no agent is using.
- **Dry Run Mode**: First show what would be deleted without actually deleting.
- **Selective Deletion**: Allow user to select which skills/versions to delete.
- **Protected Skills**: Never delete the last remaining version of a skill.

#### Scenario: Preview Global Library Cleanup
- **WHEN** User clicks "Clean Global Library" → "Preview"
- **THEN** Show list of unused skills with version counts
- **THEN** Show list of orphaned versions (not pointed to by any agent)
- **THEN** Do NOT delete anything yet

#### Scenario: Execute Global Library Cleanup
- **WHEN** User selects items from preview and clicks "Delete Selected"
- **THEN** Verify selection doesn't include last version of any skill
- **THEN** Delete selected skill directories/versions
- **THEN** Show deletion report with success/failure per item

---

### Requirement: Batch Cleanup Operations
System must support combining multiple cleanup criteria for batch operations.

- **Criteria-Based**: Filter skills by usage, age, source type, name pattern.
- **Preview First**: Always show preview before execution.
- **Atomic-ish**: Each skill deletion is independent; failures don't stop others.
- **Summary Report**: Show total processed, succeeded, failed counts.

#### Scenario: Batch Clean Unused Skills
- **WHEN** User selects "Clean unused skills" with criteria "not used by any agent in 30 days"
- **THEN** System identifies matching skills
- **THEN** Show preview with all matching skills
- **WHEN** User confirms
- **THEN** Delete each matching skill (keeping at least one version)
- **THEN** Show summary report

---

### Requirement: Cleanup Result Reporting
All cleanup operations must produce detailed, understandable reports.

- **Structured Output**: Machine-readable JSON with human-readable summary.
- **Per-Item Status**: For each attempted operation, show success/failure with reason.
- **Statistics**: Show before/after counts where applicable.

#### Data: CleanResult Structure
```go
type CleanResult struct {
    TotalProcessed int
    Succeeded     int
    Failed        int
    Items         []CleanItemResult
    Errors        []string
}

type CleanItemResult struct {
    SkillName  string
    Version    string
    Action     string // "uninstalled", "deleted", "symlink_removed"
    Success    bool
    Error      string
}
```
