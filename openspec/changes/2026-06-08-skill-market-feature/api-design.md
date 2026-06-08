# API Design & Data Models

## Backend API Extensions

### 1. Skill Market APIs

#### GetMarketConfig
```go
func (a *App) GetMarketConfig() (*MarketConfig, error)
```

```typescript
interface MarketConfig {
  url: string;
  cacheEnabled: boolean;
  cacheExpiryHours: number;
  lastUpdated?: string;
}
```

#### SetMarketConfig
```go
func (a *App) SetMarketConfig(config MarketConfig) error
```

#### ScanMarket
```go
func (a *App) ScanMarket() (*ScanMarketResult, error)
```

```typescript
interface ScanMarketResult {
  totalSkills: number;
  categories: string[];
  skills: MarketSkill[];
}

interface MarketSkill {
  name: string;
  description: string;
  author?: string;
  version?: string;
  category?: string;
  tags?: string[];
  source: string; // local path or GitHub URL
}
```

#### ListMarketSkills (with category filter)
```go
func (a *App) ListMarketSkills(category string) ([]MarketSkill, error)
```

### 2. Global Skill Scanning APIs

#### ListGlobalSkillsWithAgents
```go
func (a *App) ListGlobalSkillsWithAgents() ([]GlobalSkillWithAgents, error)
```

```typescript
interface GlobalSkillWithAgents {
  name: string;
  description: string;
  installedAgents: {
    agentId: string;
    agentName: string;
    path: string;
    version?: string;
  }[];
}
```

### 3. Project Skill APIs

#### ScanProjectSkills
```go
func (a *App) ScanProjectSkills(projectPath string) ([]ProjectSkill, error)
```

```typescript
interface ProjectSkill {
  name: string;
  description: string;
  path: string;
  isSymlink: boolean;
  symlinkTarget?: string;
}
```

#### MigrateProjectSkillToLibrary
```go
func (a *App) MigrateProjectSkillToLibrary(skillPath string, projectPath string) (*MigrateResult, error)
```

```typescript
interface MigrateResult {
  success: boolean;
  libraryPath: string;
  symlinkCreated: boolean;
  error?: string;
}
```

### 4. Batch Synchronization APIs

#### BatchSyncSkills
```go
func (a *App) BatchSyncSkills(req BatchSyncRequest) (*BatchSyncResult, error)
```

```typescript
interface BatchSyncRequest {
  skillNames: string[];
  agentIds: string[];
}

interface BatchSyncResult {
  total: number;
  succeeded: number;
  failed: number;
  results: {
    skillName: string;
    agentId: string;
    success: boolean;
    error?: string;
  }[];
}
```

## Data Model Extensions

### Extended Config Model
```go
type Config struct {
  // Existing fields...
  Skillspool SkillspoolConfig `json:"skillspool" yaml:"skillspool"`

  // New: Skill Market Configuration
  SkillMarket SkillMarketConfig `json:"skillMarket" yaml:"skillMarket"`

  Agents map[string]Agent `json:"agents" yaml:"agents"`
}

type SkillMarketConfig struct {
  URL              string `json:"url" yaml:"url"`
  CacheEnabled     bool   `json:"cacheEnabled" yaml:"cacheEnabled"`
  CacheExpiryHours int    `json:"cacheExpiryHours" yaml:"cacheExpiryHours"`
  LastUpdated      string `json:"lastUpdated,omitempty" yaml:"lastUpdated,omitempty"`
}

type SkillspoolConfig struct {
  Root string `json:"root" yaml:"root"`
}
```

### Extended Agent List
```go
var DefaultAgents = map[string]Agent{
  "trae": {
    Name:             "Trae",
    SkillLocation:    ".trae/skills",
    GlobalLocation:   "~/.trae-cn/skills",
  },
  "claude": {
    Name:             "Claude Code",
    SkillLocation:    ".claude/skills",
    GlobalLocation:   "~/.claude/skills",
  },
  "cursor": {
    Name:             "Cursor",
    SkillLocation:    ".cursor/skills",
    GlobalLocation:   "~/.cursor/skills",
  },
  "windsurf": {
    Name:             "Windsurf",
    SkillLocation:    ".windsurf/skills",
    GlobalLocation:   "~/.windsurf/skills",
  },
  "openclaw": {
    Name:             "OpenClaw",
    SkillLocation:    ".openclaw/skills",
    GlobalLocation:   "~/.openclaw/skills",
  },
  "hermes": {
    Name:             "Hermes",
    SkillLocation:    ".hermes/skills",
    GlobalLocation:   "~/.hermes/skills",
  },
  "antigravity": {
    Name:             "Antigravity",
    SkillLocation:    ".antigravity/skills",
    GlobalLocation:   "~/.antigravity/skills",
  },
  "codex": {
    Name:             "Codex",
    SkillLocation:    ".codex/skills",
    GlobalLocation:   "~/.codex/skills",
  },
  "opencode": {
    Name:             "Opencode",
    SkillLocation:    ".opencode/skills",
    GlobalLocation:   "~/.opencode/skills",
  },
}
```
