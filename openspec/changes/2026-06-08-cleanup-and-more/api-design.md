# API Design: Skill Lifecycle Management

## Overview

This document describes the API extensions for skill cleanup, version management, health checking, and usage statistics.

## API Entry Point

All methods are exposed through the `API` struct in `backend/pkg/api/api.go` and bound via Wails to the frontend.

## Method Signatures

### Cleanup Operations

```go
// UninstallFromProject removes a skill symlink from a project's agent directory.
// Does NOT remove from global library.
func (a *API) UninstallFromProject(skillName, agentID, projectPath string) error

// CleanOrphanedSymlinks finds and removes broken symlinks in an agent's skill directory.
func (a *API) CleanOrphanedSymlinks(agentID string) (int, error)
// Returns: number of symlinks cleaned

// CleanGlobalLibrary removes unused skills and orphaned versions.
func (a *API) CleanGlobalLibrary(dryRun bool) (*models.CleanResult, error)
// If dryRun=true, returns preview without deleting

// BatchCleanCriteria defines filters for batch cleanup
type BatchCleanCriteria struct {
    Unused        bool      // Skills not installed by any agent
    OlderThanDays int       // Versions older than N days
    NamePattern   string    // Glob pattern for skill names
    SourceType    string    // "github", "local", etc.
}

// BatchClean performs cleanup based on criteria
func (a *API) BatchClean(criteria models.BatchCleanCriteria) (*models.BatchCleanResult, error)
```

### Version Management

```go
// ListSkillVersions returns all installed versions of a skill
func (a *API) ListSkillVersions(skillName string) ([]models.VersionInfo, error)

type VersionInfo struct {
    Version    string
    Installed  time.Time
    SizeBytes  int64
    Source     string
    IsLatest   bool
    AgentCount int // How many agents use this version
}

// SwitchSkillVersion changes the latest pointer and updates all agent symlinks
func (a *API) SwitchSkillVersion(skillName, version string) error

// DeleteSkillVersion removes a specific version (fails if only version)
func (a *API) DeleteSkillVersion(skillName, version string) error

// CompareVersions returns detailed comparison between two versions
func (a *API) CompareVersions(skillName, v1, v2 string) (*models.VersionCompare, error)

type VersionCompare struct {
    SkillName   string
    Version1    models.VersionInfo
    Version2    models.VersionInfo
    Differences []VersionDiff
}

type VersionDiff struct {
    Field    string
    Value1   string
    Value2   string
}

// RollbackSkillVersion reverts to an older version, creating backup of current
func (a *API) RollbackSkillVersion(skillName, targetVersion string) error
```

### Health Check

```go
// CheckHealth runs all health checks and returns a comprehensive report
func (a *API) CheckHealth() (*models.HealthReport, error)

type HealthReport struct {
    GeneratedAt time.Time
    Status      string // "healthy", "warning", "error"
    Summary     HealthSummary
    Issues      []HealthIssue
    Symlinks    []SymlinkIssue
    Files       []FileIssue
    Versions    []VersionIssue
}

type HealthSummary struct {
    TotalSkills       int
    TotalAgents       int
    BrokenSymlinks    int
    MissingFiles      int
    UnreachableSkills int
}

type HealthIssue struct {
    Type         string // "broken_symlink", "missing_file", etc.
    Severity     string // "error", "warning", "info"
    SkillName    string
    AgentID      string
    Path         string
    Message      string
    Remediation  string
}

// FixBrokenSymlinks removes all broken symlinks for an agent (or all if agentID="")
func (a *API) FixBrokenSymlinks(agentID string) (int, error)

// FixLatestPointer repairs the latest pointer for a skill
func (a *API) FixLatestPointer(skillName string) error
```

### Statistics

```go
// GetSkillStats returns detailed statistics for a specific skill
func (a *API) GetSkillStats(skillName string) (*models.SkillStats, error)

type SkillStats struct {
    Name          string
    VersionCount  int
    CurrentVersion string
    TotalSizeBytes int64
    InstalledBy   []AgentInstall
    RecentActivity []ActivityEntry
}

type AgentInstall struct {
    AgentID      string
    Version      string
    InstalledAt  time.Time
}

// GetAgentStats returns statistics for a specific agent
func (a *API) GetAgentStats(agentID string) (*models.AgentStats, error)

type AgentStats struct {
    ID               string
    SkillCount       int
    OrphanedCount     int
    TotalSizeBytes   int64
    InstalledSkills  []SkillInstall
}

type SkillInstall struct {
    SkillName   string
    Version     string
    IsLatest    bool
    Status      string // "current", "update_available"
}

// GetUsageDashboard returns aggregate statistics
func (a *API) GetUsageDashboard() (*models.UsageDashboard, error)

type UsageDashboard struct {
    TotalSkills        int
    TotalInstallations int
    TotalSizeBytes     int64
    AveragePerAgent    float64
    MostPopular        []SkillCount
    LeastUsed          []SkillCount
    RecentlyActive     []SkillActivity
}

type SkillCount struct {
    SkillName string
    Count     int
}

type SkillActivity struct {
    SkillName   string
    LastActivity time.Time
}

// GetActivityTimeline returns recent activity for a skill
func (a *API) GetActivityTimeline(skillName string, days int) ([]models.ActivityEntry, error)

type ActivityEntry struct {
    Timestamp time.Time
    SkillName string
    AgentID   string
    Action    string // "installed", "uninstalled", "upgraded", "downgraded"
    Version   string
    Details   string
}

// ExportStats exports statistics in specified format
func (a *API) ExportStats(format string) ([]byte, error)
// format: "json" or "csv"
```

## Wails Binding Notes

All these methods need to be added to `app.go` to be exposed to the frontend via Wails bindings.

Example in `app.go`:
```go
// Cleanup
func (a *App) UninstallFromProject(skillName, agentID, projectPath string) error {
    return a.api.UninstallFromProject(skillName, agentID, projectPath)
}

func (a *App) CleanOrphanedSymlinks(agentID string) (int, error) {
    return a.api.CleanOrphanedSymlinks(agentID)
}

// Version Management
func (a *App) ListSkillVersions(skillName string) ([]models.VersionInfo, error) {
    return a.api.ListSkillVersions(skillName)
}

func (a *App) SwitchSkillVersion(skillName, version string) error {
    return a.api.SwitchSkillVersion(skillName, version)
}

// Health
func (a *App) CheckHealth() (*models.HealthReport, error) {
    return a.api.CheckHealth()
}

// Statistics
func (a *App) GetUsageDashboard() (*models.UsageDashboard, error) {
    return a.api.GetUsageDashboard()
}
```

## Frontend Integration

The TypeScript bindings generated by Wails will create corresponding interfaces:

```typescript
interface CleanResult {
    totalProcessed: number;
    succeeded: number;
    failed: number;
    items: CleanItemResult[];
}

interface VersionInfo {
    version: string;
    installed: string; // ISO date
    sizeBytes: number;
    source: string;
    isLatest: boolean;
    agentCount: number;
}

interface HealthReport {
    generatedAt: string;
    status: "healthy" | "warning" | "error";
    summary: HealthSummary;
    issues: HealthIssue[];
}

interface UsageDashboard {
    totalSkills: number;
    totalInstallations: number;
    totalSizeBytes: number;
    averagePerAgent: number;
    mostPopular: SkillCount[];
}
```
