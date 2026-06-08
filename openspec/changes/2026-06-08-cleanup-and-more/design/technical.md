# Design: Skill Lifecycle Management

## Design Goals

1. **Safe by default**: No accidental deletions; explicit confirmation for destructive operations
2. **Observable**: Clear feedback on what was cleaned/deleted
3. **Composable**: Can combine operations (e.g., uninstall + clean orphans in one flow)
4. **Performant**: Fast scanning for orphaned symlinks across many projects

## Technical Approach

### 1. Cleanup Operations

#### Uninstall from Project
- Remove skill symlink from project's `.agent/skills/` directory
- **Do NOT** delete from global library (skill remains available for other projects)
- Validate symlink exists before attempting removal

#### Clean Orphaned Symlinks
- Scan target directory for all symlinks
- Check if symlink target exists and is valid
- Remove broken symlinks
- Report count of cleaned symlinks

#### Global Library Cleanup
- Identify skills with zero agent installations
- Identify unreferenced versions (no agent points to them)
- Provide selective deletion with preview

### 2. Version Management

#### Storage Strategy
- Skills stored as: `{skillspool}/{skill-name}/{version}/`
- `latest` file points to current default version
- Registry tracks which version each agent is using

#### Operations
- `ListVersions(skillName)` → returns all installed versions
- `SwitchVersion(skillName, version)` → updates latest pointer + agent symlinks
- `DeleteVersion(skillName, version)` → removes version directory (fails if only version)

### 3. Health Check

#### Check Types
| Check | Method | Severity |
|-------|--------|----------|
| Broken symlinks | `os.Lstat` + `os.Stat` mismatch | Error |
| Missing files | `os.Stat` on skill files | Error |
| Version conflicts | Multiple `latest` files | Warning |
| Unreachable skills | Registry entry but no directory | Warning |

#### Output Format
```json
{
  "healthy": true,
  "errors": [],
  "warnings": [
    {"type": "unreachable", "skill": "xyz", "detail": "registry entry but no directory"}
  ]
}
```

### 4. Usage Statistics

#### Data Tracked
- Per skill: list of agents that have installed it
- Per agent: list of installed skills (from symlinks)
- Skill source (GitHub URL, local path, etc.)

#### Display Format
- Skills tab: show badge with agent count
- Agent tab: show skill count per agent

## API Surface

### New API Methods

```go
// Cleanup
UninstallFromProject(skillName, agentID, projectPath) error
CleanOrphanedSymlinks(agentID) (int, error)
CleanGlobalLibrary(dryRun bool) (*CleanResult, error)
BatchClean(criteria BatchCleanCriteria) (*BatchCleanResult, error)

// Version Management
ListSkillVersions(skillName) ([]string, error)
SwitchSkillVersion(skillName, version) error
DeleteSkillVersion(skillName, version) error

// Health & Stats
CheckHealth() (*HealthReport, error)
GetSkillStats(skillName) (*SkillStats, error)
GetAgentStats(agentID) (*AgentStats, error)
```

## Data Model Extensions

```go
// SkillStats tracks usage
type SkillStats struct {
    Name        string
    VersionCount int
    AgentIDs    []string
    TotalSize   int64 // bytes
}

// AgentStats tracks agent health
type AgentStats struct {
    ID              string
    SkillCount      int
    OrphanedCount    int
    InstalledSkills []string
}

// HealthReport contains check results
type HealthReport struct {
    Healthy   bool
    Errors    []HealthIssue
    Warnings  []HealthIssue
}
```
