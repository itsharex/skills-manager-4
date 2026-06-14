# Usage Statistics Specification

## Purpose
Provide visibility into skill usage patterns across the agent ecosystem.

---

## Requirements

### Requirement: Skill Usage Overview
System must show which agents have installed each skill.

- **Agent List Per Skill**: For each skill, list all agents that have it installed.
- **Installation Type**: Indicate if agent has latest version or a specific version.
- **Source Information**: Show where the skill originally came from (GitHub URL, local).
- **Sort Options**: Sort skills by name, agent count, size, last used.

#### Scenario: View Skill Usage
- **WHEN** User selects a skill and clicks "Usage" or hovers to see popover
- **THEN** Display usage panel:
  ```
  Skill: using-superpowers
  Installed by: 4 agents
  
  ┌──────────┬─────────┬──────────────┐
  │ Agent    │ Version │ Source       │
  ├──────────┼─────────┼──────────────┤
  │ Trae     │ v1.2.0  │ GitHub (⭐)  │
  │ Cursor   │ v1.2.0  │ GitHub       │
  │ Claude   │ v1.1.0  │ GitHub       │
  │ Windsurf │ latest  │ Local        │
  └──────────┴─────────┴──────────────┘
  ```

---

### Requirement: Agent Skill Inventory
System must show all skills installed by a specific agent.

- **Complete List**: List all skills (symlinks) in agent's skill directory.
- **Version Display**: Show which version of each skill the agent is using.
- **Status Indicators**: Show if skill is "up to date" or "update available".
- **Skill Count**: Total count of skills installed by agent.

#### Scenario: View Agent Skill Inventory
- **WHEN** User navigates to Agent detail and selects "Skills" tab
- **THEN** Display skill inventory:
  ```
  Agent: Trae
  Skills installed: 15
  
  ┌────────────────────┬─────────┬────────────┐
  │ Skill              │ Version │ Status     │
  ├────────────────────┼─────────┼────────────┤
  │ using-superpowers  │ v1.2.0  │ ✅ Current │
  │ code-review        │ v2.0.1  │ ⚠️ Update  │
  │ github-tools       │ latest  │ ✅ Current │
  └────────────────────┴─────────┴────────────┘
  ```

---

### Requirement: Usage Statistics Dashboard
System must provide aggregate statistics across the entire ecosystem.

- **Total Skills**: Count of unique skills in global library.
- **Total Installations**: Sum of all skill installations across all agents.
- **Average per Agent**: Total installations / number of agents.
- **Most Popular**: Top 5 most-installed skills.
- **Least Used**: Skills installed by fewest agents.

#### Scenario: View Usage Dashboard
- **WHEN** User navigates to "Statistics" or "Dashboard"
- **THEN** Display aggregate stats:
  ```
  ┌─────────────────────────────────────────┐
  │ 📊 Skills Statistics                     │
  ├─────────────────────────────────────────┤
  │ Total unique skills: 24                 │
  │ Total installations: 87                 │
  │ Average per agent: 10.9                 │
  │                                         │
  │ Most Popular:                           │
  │   1. using-superpowers (4 agents)       │
  │   2. github-tools (3 agents)             │
  │   3. code-review (3 agents)             │
  │                                         │
  │ Least Used:                             │
  │   1. legacy-migration (1 agent)         │
  │   2. old-api-adapter (1 agent)          │
  └─────────────────────────────────────────┘
  ```

---

### Requirement: Skill Size Analytics
System must track and display storage usage per skill and total.

- **Per-Skill Size**: Calculate total bytes used by each skill (all versions).
- **Version Breakdown**: Show size per version within a skill.
- **Total Ecosystem Size**: Sum of all skills in skillspool.
- **Agent Contribution**: Show how much space each agent's symlinks represent (none, symlinks are tiny).

#### Scenario: View Skill Sizes
- **WHEN** User views skill detail or size column
- **THEN** Display size information:
  ```
  Skill: using-superpowers
  Total size: 15.2 MB (across 3 versions)
  
  Versions:
    v1.0.0 - 4.8 MB (deleted)
    v1.1.0 - 5.1 MB 
    v1.2.0 - 5.3 MB (current)
  ```

---

### Requirement: Activity Timeline
System must track and display recent skill activity.

- **Installation Log**: Record each skill installation with timestamp.
- **Uninstall Log**: Record each uninstall with timestamp.
- **Version Change Log**: Record when agents switch versions.
- **Retention Period**: Keep logs for 30 days by default.

#### Scenario: View Activity Timeline
- **WHEN** User views skill detail and selects "Activity" tab
- **THEN** Display timeline:
  ```
  Recent Activity for: using-superpowers
  
  • 2024-02-15 14:32 - Trae upgraded to v1.2.0
  • 2024-02-14 09:15 - Claude installed v1.1.0
  • 2024-02-10 16:45 - Cursor uninstalled
  • 2024-02-01 11:20 - Windsurf installed latest
  ```

---

### Requirement: Export Statistics
System must allow exporting usage data for external analysis.

- **CSV Export**: Tabular format for spreadsheet analysis.
- **JSON Export**: Full structured data for programmatic access.
- **Date Range**: Filter exports by date range.
- **Include Fields**: Skill name, agents, versions, sizes, activity.

#### Data: UsageStats Structure
```go
type UsageStats struct {
    GeneratedAt time.Time
    
    // Aggregate
    TotalSkills        int
    TotalInstallations int
    TotalSizeBytes     int64
    
    // Per-skill
    Skills []SkillStats
    
    // Activity (last 30 days)
    RecentActivity []ActivityEntry
}

type SkillStats struct {
    Name           string
    AgentCount     int
    VersionCount   int
    CurrentVersion string
    SizeBytes      int64
    InstalledBy    []AgentInstall
}

type AgentInstall struct {
    AgentID   string
    Version   string
    Installed time.Time
}

type ActivityEntry struct {
    Timestamp   time.Time
    SkillName   string
    AgentID     string
    Action      string // "installed", "uninstalled", "upgraded", "downgraded"
    Version     string
    Details     string
}
```
