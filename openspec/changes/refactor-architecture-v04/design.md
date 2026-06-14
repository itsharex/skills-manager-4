## Context

Current architecture conflates skill pool (storage) with skill scanning (discovery). Market is just a raw URL input. Agent detection checks wrong paths. This design addresses all four.

## Goals / Non-Goals

### Goals
- 技能池 = 可配置的本地技能仓库目录
- 市场 = 可配置的多来源搜索（本地池优先 → GitHub → 开放市场）
- 扫描 = 独立于池的机器扫描能力，分"已收录/未收录"
- Agent 检测 = 检测 agent 二进制/应用本身安装

### Non-Goals
- 不做远程 registry 的完整实现（只做 URL 格式支持）
- 不做 GitHub 同步的定时任务机制（只做手动同步）
- 不做技能订阅/自动更新

## Decisions

### Decision 1: Config 扩展

```go
type Config struct {
    RepoPath      string         // ~/.skill-repo/ (unchanged)
    PoolPath      string         // NEW: 本地技能池目录，默认 ~/.skill-pool/
    InstallMode   string         // "symlink" | "copy"
    AutoFallback  bool
    DefaultAgents []string
    MarketSources []MarketSource // NEW: 市场来源列表
    LinkTargets   []LinkTarget
    Repositories  []RepoSource   // kept for backward compat
    CacheTTL      int
}

type MarketSource struct {
    Name    string // 显示名
    URL     string // 本地路径 / GitHub URL / registry URL
    Type    string // "pool" | "github" | "registry"
    Enabled bool
    Branch  string // GitHub branch
}
```

**Alternatives considered:**
- 将 MarketSources 独立为单独文件？→ 过于复杂，保持 Config 统一管理
- 用 `Repositories` 复用？→ 类型语义不同，新增更清晰

### Decision 2: Agent 检测重写

```go
type Agent struct {
    ID          string
    Name        string
    SkillsDir   string  // 始终返回（无论是否检测到）
    DetectCmd   string  // CLI 检测命令名，如 "codex"（可选）
    DetectPath  string  // 目录/文件检测路径，如 ~/.cursor/（可选）
    AutoDetected bool
}

func DetectAgents() []Agent {
    for _, agent := range KnownAgents() {
        detected := false
        if agent.DetectCmd != "" {
            _, err := exec.LookPath(agent.DetectCmd)
            detected = err == nil
        } else if agent.DetectPath != "" {
            _, err := os.Stat(agent.DetectPath)
            detected = err == nil
        }
        agent.AutoDetected = detected
    }
}
```

**Why not check SkillsDir?** SkillsDir 只在安装技能时才被创建，检测它永远返回 false。
**Why have both DetectCmd and DetectPath?** CLI 工具用 `LookPath`，IDE/桌面应用用 `os.Stat`。

### Decision 3: 池 vs 扫描分离

```
SkillPoolPage:
  1. 读取 config.PoolPath 目录下的技能子目录 → 显示池内容
  2. "本机扫描" 按钮 → ScanLocal() → 匹配池 → 分已收录/未收录

ScanLocal(projectPath):
  1. 扫描所有 agent SkillsDir
  2. 若指定 projectPath，额外扫描项目下的 agent 子目录
  3. 返回结果 + 池匹配标记 (alreadyInPool)
```

### Decision 4: 市场搜索优先级

```
MarketPage:
  1. 读取 config.MarketSources
  2. 显示来源列表，每个可单独搜索/禁用
  3. "搜索所有" → 按优先级: pool(本地目录) → github → registry
  4. 结果合并显示，标记来源名
```

## Risks / Trade-offs

- **Config 向后兼容**: `PoolPath` 默认空串，旧配置加载后自动用 `~/.skill-pool/` 替代
- **MarketSources 为空**: 市场页面显示"请先在设置中添加市场来源"提示
- **Agent 检测跨平台**: `exec.LookPath` 在 Windows/macOS/Linux 均可用；检测路径需各平台适配

## Migration Plan

1. `LoadConfig` 中如果 `PoolPath == ""`，设置为 `filepath.Join(home, ".skill-pool")`
2. 旧 `ScanPool` 方法标记为 deprecated，逐步替换

## Open Questions

- 市场来源的 `github` 类型如何实现"搜索"？用 `gh` CLI 还是直接 GitHub API？
- 默认自带哪些市场来源？