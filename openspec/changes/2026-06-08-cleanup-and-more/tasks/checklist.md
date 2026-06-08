# Tasks: Implementation Checklist

## 1. Data Models & Types
- [ ] Extend `models.go` with new types:
  - [ ] `CleanResult` / `CleanItemResult`
  - [ ] `VersionStats`
  - [ ] `HealthReport` / `HealthIssue`
  - [ ] `SkillStats` / `AgentStats` / `UsageStats`
  - [ ] `ActivityEntry`

## 2. Backend API Implementation
### 2.1 Cleanup APIs
- [ ] `UninstallFromProject(skillName, agentID, projectPath) error`
- [ ] `CleanOrphanedSymlinks(agentID string) (int, error)`
- [ ] `CleanGlobalLibrary(dryRun bool) (*CleanResult, error)`
- [ ] `BatchClean(criteria BatchCleanCriteria) (*BatchCleanResult, error)`

### 2.2 Version Management APIs
- [ ] `ListSkillVersions(skillName string) ([]string, error)`
- [ ] `SwitchSkillVersion(skillName, version string) error`
- [ ] `DeleteSkillVersion(skillName, version string) error`
- [ ] `CompareVersions(skillName, v1, v2 string) (*VersionCompare, error)`
- [ ] `RollbackSkillVersion(skillName, targetVersion string) error`

### 2.3 Health Check APIs
- [ ] `CheckHealth() (*HealthReport, error)`
- [ ] `FixBrokenSymlinks(agentID string) (int, error)`
- [ ] `FixLatestPointer(skillName string) error`
- [ ] `RemoveUnreachableSkill(skillName string) error`

### 2.4 Statistics APIs
- [ ] `GetSkillStats(skillName string) (*SkillStats, error)`
- [ ] `GetAgentStats(agentID string) (*AgentStats, error)`
- [ ] `GetUsageDashboard() (*UsageStats, error)`
- [ ] `GetActivityTimeline(skillName string, days int) ([]ActivityEntry, error)`
- [ ] `ExportStats(format string) ([]byte, error)`

## 3. Wails Bindings
- [ ] Add all new API methods to `app.go` for Wails binding
- [ ] Ensure all new types are exported for JSON serialization
- [ ] Test `wails dev` generates correct TypeScript bindings

## 4. Frontend Implementation
### 4.1 UI Components
- [ ] Health Dashboard component with status cards
- [ ] Cleanup wizard/modal components
- [ ] Version list component with actions
- [ ] Version comparison modal
- [ ] Statistics charts/tables
- [ ] Activity timeline component

### 4.2 Pages
- [ ] Update SkillsPage with cleanup actions
- [ ] Update AgentsPage with stats and cleanup
- [ ] New "Health" page with overview
- [ ] New "Statistics" page with dashboard

### 4.3 Design
- [ ] Apply fresh/clean theme per frontend-design skill
- [ ] Ensure consistent spacing, typography
- [ ] Add appropriate animations/micro-interactions
- [ ] Responsive design for all pages

## 5. Testing
- [ ] Backend unit tests for cleanup logic
- [ ] Backend unit tests for health checks
- [ ] Backend unit tests for version operations
- [ ] Manual testing of all new features
- [ ] Build verification

## 6. Documentation
- [ ] Update README with new features
- [ ] Update openspec if needed
- [ ] CLI help text for new commands
