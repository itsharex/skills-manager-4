# Health Check Specification

## Purpose
Provide comprehensive health monitoring and diagnostics for the skills ecosystem.

---

## Requirements

### Requirement: System Health Overview
System must provide a dashboard-level view of overall skills ecosystem health.

- **Overall Status**: Single status indicator (Healthy/Warning/Error) for the entire system.
- **Summary Counts**: Count of skills, agents, broken symlinks, missing files.
- **Quick Issues**: List of top 5 issues requiring attention.
- **Last Check Time**: Timestamp of most recent health check.

#### Scenario: View Health Overview
- **WHEN** User navigates to "Health" tab or dashboard widget
- **THEN** Display health overview card:
  ```
  ┌─────────────────────────────────────┐
  │ 🟢 System Healthy                    │
  │ Last checked: 2 minutes ago         │
  │                                     │
  │ Skills: 24 | Agents: 8             │
  │ Broken symlinks: 2 | Errors: 0      │
  │                                     │
  │ ⚠️ 2 warnings:                      │
  │   • 2 orphaned symlinks in Cursor   │
  │   • 1 unreachable skill in Trae     │
  └─────────────────────────────────────┘
  ```

---

### Requirement: Broken Symlink Detection
System must detect symlinks that point to non-existent targets.

- **Symlink Enumeration**: Scan all agent skill directories for symlinks.
- **Target Validation**: Check if symlink target exists using `os.Stat()`.
- **Broken Link Identification**: Links where target doesn't exist are "broken".
- **Agent Attribution**: Associate broken symlinks with specific agents.

#### Scenario: Detect Broken Symlinks
- **WHEN** Health check runs
- **FOR EACH** agent's skill directory:
  - Scan for symlinks (files where `Mode() & os.ModeSymlink != 0`)
  - For each symlink, call `os.Stat()` to check target
  - If `os.Stat()` fails, mark as broken
- **THEN** Report broken symlinks grouped by agent

#### Detection Logic
```go
func isBrokenSymlink(path string) bool {
    link, err := os.Readlink(path)
    if err != nil {
        return true // can't read link
    }
    // Resolve relative links
    if !filepath.IsAbs(link) {
        link = filepath.Join(filepath.Dir(path), link)
    }
    _, err = os.Stat(link)
    return err != nil // target doesn't exist
}
```

---

### Requirement: Missing File Detection
System must detect skills with missing or corrupted files.

- **Required Files**: Each skill must have `SKILL.md` at minimum.
- **File Enumeration**: Scan skill directory for all registered files.
- **Existence Check**: Verify each file still exists on disk.
- **Size Validation**: Flag files with zero size as potentially corrupted.

#### Scenario: Detect Missing Skill Files
- **WHEN** Health check runs for a skill
- **FOR EACH** file registered in the skill's registry entry:
  - Call `os.Stat()` to verify existence
  - If fails, add to missing files list
- **THEN** If `SKILL.md` is missing, mark skill as "corrupted" (critical)

---

### Requirement: Version Conflict Detection
System must detect when multiple versions claim to be "latest".

- **Latest File Uniqueness**: Only one version should have the `latest` file pointing to it.
- **Circular Latest**: Detect if `latest` file points to non-existent version.
- **Orphan Latest**: Detect if `latest` file references a version directory that doesn't exist.

#### Scenario: Detect Version Conflicts
- **WHEN** Health check runs
- **FOR EACH** skill in registry:
  - Read `latest` file content (should be one version dir name)
  - Check if that version directory exists
  - If `latest` points to missing dir, mark as conflict
- **IF** multiple `latest` markers exist (shouldn't happen but check anyway), flag as error

---

### Requirement: Unreachable Skill Detection
System must detect skills registered in the database but with no actual directory.

- **Registry Scan**: Iterate all skills in registry.
- **Directory Check**: Verify `{skillspool}/{skill-name}/` exists.
- **Missing Directory**: Skills with no directory are "unreachable".

#### Scenario: Detect Unreachable Skills
- **WHEN** Health check runs
- **FOR EACH** skill in registry:
  - Check if `{skillspool}/{skill.Name}/` directory exists
  - If not, add to unreachable list
- **THEN** These skills need either restoration or registry cleanup

---

### Requirement: Health Check Scheduling
System must support both manual and automatic health checks.

- **Manual Trigger**: User can click "Run Health Check" at any time.
- **Auto-trigger**: Health check runs automatically on app startup (background).
- **Refresh Interval**: Configurable auto-check interval (default: every 30 minutes).
- **Stale Warning**: If last check > 1 hour, show "stale data" warning.

#### Scenario: Manual Health Check
- **WHEN** User clicks "Run Health Check"
- **THEN** Show "Checking..." spinner overlay
- **THEN** Run all health checks in parallel where possible
- **THEN** Update UI with results
- **THEN** Update "Last checked" timestamp

---

### Requirement: Health Issue Remediation
System must provide one-click or guided remediation for detected issues.

| Issue Type | Auto-fix Available? | Guidance |
|------------|-------------------|----------|
| Broken symlink | ✅ Yes | "Clean orphaned symlinks" button |
| Missing files | ❌ No | "Reinstall skill" button |
| Version conflict | ✅ Yes | "Fix latest pointer" button |
| Unreachable skill | ❌ No | "Remove from registry" or "Restore" |

#### Scenario: Auto-fix Broken Symlinks
- **WHEN** User views health issues and sees broken symlinks
- **THEN** Show "Clean Orphaned Symlinks" button
- **WHEN** User clicks button
- **THEN** Execute cleanup operation
- **THEN** Refresh health check and update display

---

### Requirement: Health Report Export
System must allow exporting health reports for debugging/support.

- **JSON Export**: Full machine-readable report.
- **Summary Export**: Human-readable text/markdown summary.
- **Include Metadata**: Report should include system info, timestamps, versions.

#### Data: HealthReport Structure
```go
type HealthReport struct {
    GeneratedAt   time.Time
    SystemInfo    SystemInfo // OS, app version, etc.
    OverallStatus Status     // "healthy", "warning", "error"
    Summary       HealthSummary
    Issues        []HealthIssue
    
    // Per-category breakdowns
    SymlinkIssues []SymlinkIssue
    FileIssues    []FileIssue
    VersionIssues []VersionIssue
}

type HealthIssue struct {
    Type     IssueType
    Severity Severity // "error", "warning", "info"
    Message  string
    Skill    string
    Agent    string
    Detail   string
    Remediation string
}
```
