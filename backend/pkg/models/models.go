package models

import "time"

// SourceType 表示技能的安装来源类型
type SourceType string

const (
	SourceGitHub   SourceType = "github"
	SourceNpx      SourceType = "npx"
	SourceLocal    SourceType = "local"
	SourceRegistry SourceType = "registry"
	SourceClawhub  SourceType = "clawhub"
)

// Source 记录技能的来源信息
type Source struct {
	Type    SourceType `json:"type" yaml:"type"`
	URL     string     `json:"url,omitempty" yaml:"url,omitempty"`
	Ref     string     `json:"ref,omitempty" yaml:"ref,omitempty"`
	SubPath string     `json:"sub_path,omitempty" yaml:"sub_path,omitempty"`
	Command string     `json:"command,omitempty" yaml:"command,omitempty"`
	Path    string     `json:"path,omitempty" yaml:"path,omitempty"`
}

// SkillInfo 从 SKILL.md 解析出的技能元信息
type SkillInfo struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Version     string   `json:"version,omitempty" yaml:"version,omitempty"`
	Author      string   `json:"author,omitempty" yaml:"author,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// SkillInstallStatus 技能的安装状态
type SkillInstallStatus string

const (
	StatusNotInstalled     SkillInstallStatus = "not_installed"     // 技能池中有但未安装到任何 Agent
	StatusPartiallyInstalled SkillInstallStatus = "partially_installed" // 仅安装到部分 Agent
	StatusInstalled        SkillInstallStatus = "installed"         // 已安装到所有已检测的 Agent
	StatusProjectOnly      SkillInstallStatus = "project_only"      // 仅项目目录中有
)

// SkillVersion 已安装的技能版本信息
type SkillVersion struct {
	Version     string   `json:"version"`
	InstalledAt string   `json:"installed_at"`
	Path        string   `json:"path"`
	Agents      []string `json:"agents"`
}

// Skill 注册表中的技能条目
type Skill struct {
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	Source        Source                  `json:"source"`
	Versions      map[string]SkillVersion `json:"versions"`
	LatestVersion string                  `json:"latest_version"`
	UserTags      []string                `json:"user_tags,omitempty"` // 用户自定义标签
}

// Registry 技能注册表
type Registry struct {
	Version     int               `json:"version"`
	InstalledAt string            `json:"installed_at"`
	Skills      map[string]*Skill `json:"skills"`
}

// Agent 配置的 Agent 信息
type Agent struct {
	Name                string `json:"name" yaml:"name"`
	SkillLocation       string `json:"skill_location" yaml:"skill_location"`                 // 项目级技能目录（相对项目根）
	GlobalLocation      string `json:"global_location" yaml:"global_location"`               // 全局技能目录（相对用户 home）
	Installed           bool   `json:"installed" yaml:"installed"`
	Detected            bool   `json:"detected" yaml:"detected"`
	SupportsProject     bool   `json:"supports_project" yaml:"supports_project"`             // 是否支持项目级安装
	GlobalDirectoryKey  string `json:"global_directory_key" yaml:"global_directory_key"`     // 全局目录标识（用于检测多 Agent 共享目录）
	ProjectDirectoryKey string `json:"project_directory_key" yaml:"project_directory_key"`   // 项目级目录标识（用于检测多 Agent 共享目录）
}

// AgentGroup 表示共享同一技能目录的 Agent 组
type AgentGroup struct {
	ID          string   `json:"id"`           // 目录标识（如 ".agents/skills"）
	Directory   string   `json:"directory"`    // 实际目录路径（带 ~/ 或相对）
	AgentIDs    []string `json:"agentIds"`     // 该目录下的所有 Agent
	AgentNames  []string `json:"agentNames"`   // 便于前端展示
	DetectedIDs []string `json:"detectedIds"`  // 已检测到的 Agent
	Scope       string   `json:"scope"`        // "global" 或 "project"
	SharedRisk  bool     `json:"sharedRisk"`   // 是否有共享目录风险（>=2 个 Agent）
}

// SkillspoolConfig 技能池配置
type SkillspoolConfig struct {
	Root string `json:"root" yaml:"root"`
}

// SkillspoolMigrationResult 描述 skillspool 路径迁移结果
type SkillspoolMigrationResult struct {
	Success    bool   `json:"success"`
	OldRoot    string `json:"old_root"`
	NewRoot    string `json:"new_root"`
	MovedFiles int    `json:"moved_files"`
	Message    string `json:"message"`
}

// SkillMarketConfig 技能市场配置
type SkillMarketConfig struct {
	URL              string `json:"url" yaml:"url"`
	CacheEnabled     bool   `json:"cacheEnabled" yaml:"cacheEnabled"`
	CacheExpiryHours int    `json:"cacheExpiryHours" yaml:"cacheExpiryHours"`
	LastUpdated      string `json:"lastUpdated,omitempty" yaml:"lastUpdated,omitempty"`
}

// MarketSkill 技能市场中的技能
type MarketSkill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author,omitempty"`
	Version     string   `json:"version,omitempty"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Source      string   `json:"source"`
}

// ScanMarketResult 市场扫描结果
type ScanMarketResult struct {
	TotalSkills int           `json:"totalSkills"`
	Categories  []string      `json:"categories"`
	Skills      []MarketSkill `json:"skills"`
}

// GlobalSkillWithAgents 带有 Agent 安装信息的全局技能
type GlobalSkillWithAgents struct {
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	InstalledAgents []AgentSkillStatus   `json:"installedAgents"`
}

// AgentSkillStatus 技能在 Agent 中的安装状态
type AgentSkillStatus struct {
	AgentID   string `json:"agentId"`
	AgentName string `json:"agentName"`
	Path      string `json:"path"`
	Version   string `json:"version,omitempty"`
}

// ProjectSkill 项目中的技能
type ProjectSkill struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Path          string `json:"path"`
	IsSymlink     bool   `json:"isSymlink"`
	SymlinkTarget string `json:"symlinkTarget,omitempty"`
	// 新增字段：是否已存在于全局技能池
	InLibrary bool `json:"inLibrary"`
	// 版本号
	Version string `json:"version,omitempty"`
	// 标签（从 SKILL.md 解析）
	Tags []string `json:"tags,omitempty"`
	// 大小（字节）
	SizeBytes int64 `json:"sizeBytes,omitempty"`
}

// SkillWithStatus 带状态信息的技能（用于前端展示）
type SkillWithStatus struct {
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Tags            []string           `json:"tags,omitempty"`
	LatestVersion   string             `json:"latestVersion"`
	InstallStatus   SkillInstallStatus `json:"installStatus"`
	InstalledAgents []string           `json:"installedAgents"`
	TotalAgents     int                `json:"totalAgents"`
	Source          Source             `json:"source"`
	SizeBytes       int64              `json:"sizeBytes,omitempty"`
}

// SkillListFilter 技能列表过滤条件
type SkillListFilter struct {
	Status   string   `json:"status,omitempty"`   // "", "installed", "not_installed", "partially_installed", "project_only"
	Tag      string   `json:"tag,omitempty"`      // 按标签过滤
	Search   string   `json:"search,omitempty"`   // 关键词搜索
	AgentIDs []string `json:"agentIds,omitempty"` // 按使用的 Agent 过滤
}

// TagUsage 标签使用情况
type TagUsage struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// MigrateResult 迁移结果
type MigrateResult struct {
	Success        bool   `json:"success"`
	LibraryPath    string `json:"libraryPath"`
	SymlinkCreated bool   `json:"symlinkCreated"`
	Error          string `json:"error,omitempty"`
}

// BatchSyncRequest 批量同步请求
type BatchSyncRequest struct {
	SkillNames []string `json:"skillNames"`
	AgentIDs   []string `json:"agentIds"`
}

// BatchSyncResult 批量同步结果
type BatchSyncResult struct {
	Total     int                   `json:"total"`
	Succeeded int                   `json:"succeeded"`
	Failed    int                   `json:"failed"`
	Results   []BatchSyncItemResult `json:"results"`
}

// BatchSyncItemResult 单条同步结果
type BatchSyncItemResult struct {
	SkillName string `json:"skillName"`
	AgentID   string `json:"agentId"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// Config 应用配置
type Config struct {
	Skillspool  SkillspoolConfig  `json:"skillspool" yaml:"skillspool"`
	SkillMarket SkillMarketConfig `json:"skillMarket" yaml:"skillMarket"`
	Agents      map[string]Agent  `json:"agents" yaml:"agents"`
}

// --- 清理功能相关类型 ---

// CleanResult 清理操作结果
type CleanResult struct {
	TotalProcessed int                `json:"totalProcessed"`
	Succeeded      int                `json:"succeeded"`
	Failed         int                `json:"failed"`
	Items          []CleanItemResult  `json:"items"`
	Errors         []string           `json:"errors,omitempty"`
}

// CleanItemResult 单个清理项的结果
type CleanItemResult struct {
	SkillName string `json:"skillName"`
	Version   string `json:"version,omitempty"`
	Action    string `json:"action"` // "uninstalled", "deleted", "symlink_removed"
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// BatchCleanCriteria 批量清理条件
type BatchCleanCriteria struct {
	Unused        bool   `json:"unused"`         // 未安装到任何 Agent 的技能
	OlderThanDays int    `json:"olderThanDays"`  // 早于 N 天前安装
	NamePattern   string `json:"namePattern"`    // 技能名称匹配模式
	SourceType    string `json:"sourceType"`     // 来源类型: github, local
}

// BatchCleanResult 批量清理结果
type BatchCleanResult struct {
	Total     int                `json:"total"`
	Succeeded int                `json:"succeeded"`
	Failed    int                `json:"failed"`
	Results   []CleanItemResult  `json:"results"`
	DryRun    bool               `json:"dryRun"`
}

// --- 版本管理相关类型 ---

// VersionInfo 版本信息
type VersionInfo struct {
	Version    string    `json:"version"`
	Installed  time.Time `json:"installed"`
	SizeBytes int64     `json:"sizeBytes"`
	Source     string    `json:"source"`
	IsLatest   bool      `json:"isLatest"`
	AgentCount int       `json:"agentCount"`
}

// VersionCompare 版本对比结果
type VersionCompare struct {
	SkillName   string        `json:"skillName"`
	Version1    VersionInfo   `json:"version1"`
	Version2    VersionInfo   `json:"version2"`
	Differences []VersionDiff `json:"differences"`
}

// VersionDiff 版本差异项
type VersionDiff struct {
	Field  string `json:"field"`
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
}

// --- 健康检查相关类型 ---

// HealthReport 健康检查报告
type HealthReport struct {
	GeneratedAt time.Time      `json:"generatedAt"`
	Status      string         `json:"status"` // "healthy", "warning", "error"
	Summary     HealthSummary `json:"summary"`
	Issues      []HealthIssue `json:"issues"`
	Symlinks    []SymlinkIssue `json:"symlinks,omitempty"`
	Files       []FileIssue   `json:"files,omitempty"`
	Versions    []VersionIssue `json:"versions,omitempty"`
}

// HealthSummary 健康检查摘要
type HealthSummary struct {
	TotalSkills        int `json:"totalSkills"`
	TotalAgents        int `json:"totalAgents"`
	BrokenSymlinks     int `json:"brokenSymlinks"`
	MissingFiles       int `json:"missingFiles"`
	UnreachableSkills  int `json:"unreachableSkills"`
}

// HealthIssue 健康问题
type HealthIssue struct {
	Type        string `json:"type"`        // "broken_symlink", "missing_file", "unreachable"
	Severity    string `json:"severity"`    // "error", "warning", "info"
	SkillName   string `json:"skillName,omitempty"`
	AgentID     string `json:"agentId,omitempty"`
	Path        string `json:"path"`
	Message     string `json:"message"`
	Remediation string `json:"remediation,omitempty"`
}

// SymlinkIssue 软链问题
type SymlinkIssue struct {
	AgentID   string `json:"agentId"`
	Path      string `json:"path"`
	Target    string `json:"target"`
	TargetExists bool `json:"targetExists"`
}

// FileIssue 文件问题
type FileIssue struct {
	SkillName string `json:"skillName"`
	Path      string `json:"path"`
	Missing   bool   `json:"missing"`
}

// VersionIssue 版本问题
type VersionIssue struct {
	SkillName   string `json:"skillName"`
	Version     string `json:"version"`
	Issue      string `json:"issue"` // "latest_missing", "latest_broken"
}

// --- 统计相关类型 ---

// SkillStats 技能统计
type SkillStats struct {
	Name           string           `json:"name"`
	VersionCount   int              `json:"versionCount"`
	CurrentVersion string           `json:"currentVersion"`
	SizeBytes     int64            `json:"sizeBytes"`
	InstalledBy   []AgentInstall   `json:"installedBy"`
}

// AgentInstall Agent 安装信息
type AgentInstall struct {
	AgentID     string    `json:"agentId"`
	Version     string    `json:"version"`
	InstalledAt time.Time `json:"installedAt"`
}

// AgentStats Agent 统计
type AgentStats struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	SkillCount      int            `json:"skillCount"`
	OrphanedCount   int            `json:"orphanedCount"`
	TotalSizeBytes  int64          `json:"totalSizeBytes"`
	InstalledSkills []SkillInstall `json:"installedSkills"`
}

// SkillInstall 技能安装信息
type SkillInstall struct {
	SkillName string `json:"skillName"`
	Version   string `json:"version"`
	IsLatest  bool   `json:"isLatest"`
	Status    string `json:"status"` // "current", "update_available"
}

// UsageDashboard 使用统计仪表盘
type UsageDashboard struct {
	TotalSkills        int            `json:"totalSkills"`
	TotalInstallations int            `json:"totalInstallations"`
	TotalSizeBytes     int64          `json:"totalSizeBytes"`
	AveragePerAgent    float64        `json:"averagePerAgent"`
	MostPopular        []SkillCount   `json:"mostPopular"`
	LeastUsed          []SkillCount   `json:"leastUsed"`
	RecentlyActive     []SkillActivity `json:"recentlyActive"`
}

// SkillCount 技能计数
type SkillCount struct {
	SkillName string `json:"skillName"`
	Count     int    `json:"count"`
}

// SkillActivity 技能最近活动
type SkillActivity struct {
	SkillName    string    `json:"skillName"`
	LastActivity time.Time `json:"lastActivity"`
}

// ActivityEntry 活动条目
type ActivityEntry struct {
	Timestamp time.Time `json:"timestamp"`
	SkillName string    `json:"skillName"`
	AgentID   string    `json:"agentId,omitempty"`
	Action    string    `json:"action"` // "installed", "uninstalled", "upgraded", "downgraded"
	Version   string    `json:"version"`
	Details   string    `json:"details,omitempty"`
}

// ---------- ClawHub 适配相关类型 ----------

// ClawHubSkill 表示 ClawHub 市场上的一个技能
type ClawHubSkill struct {
	Owner       string   `json:"owner"`       // 发布者/组织
	Slug        string   `json:"slug"`        // 技能名
	Name        string   `json:"name"`        // 显示名（owner/slug）
	Description string   `json:"description"` // 简介
	Version     string   `json:"version,omitempty"`
	Author      string   `json:"author,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Downloads   int      `json:"downloads,omitempty"`
	Stars       int      `json:"stars,omitempty"`
	UpdatedAt   string   `json:"updatedAt,omitempty"`
}

// ClawHubInstallRequest 从 ClawHub 安装技能的请求
type ClawHubInstallRequest struct {
	Owner      string   `json:"owner"`      // skill 发布者
	Slug       string   `json:"slug"`       // skill 名
	AgentIDs   []string `json:"agentIds"`   // 目标 Agent
	ProjectDir string   `json:"projectDir"` // 项目级安装时的项目目录（可选）
	Scope      string   `json:"scope"`      // "global" 或 "project"
}

// RuntimeStatus Node.js / clawhub CLI 运行时状态
type RuntimeStatus struct {
	NodeInstalled     bool   `json:"nodeInstalled"`
	NodeVersion       string `json:"nodeVersion,omitempty"`
	NodePath          string `json:"nodePath,omitempty"`
	ClawHubInstalled  bool   `json:"clawhubInstalled"`
	ClawHubVersion    string `json:"clawhubVersion,omitempty"`
	ClawHubPath       string `json:"clawhubPath,omitempty"`
	HasNpm            bool   `json:"hasNpm"`
	Message           string `json:"message,omitempty"`
	RegistryReachable bool   `json:"registryReachable,omitempty"`
	RegistryName      string `json:"registryName,omitempty"`
}

// RuntimeInstallReport 运行时安装报告
type RuntimeInstallReport struct {
	Success bool   `json:"success"`
	Step    string `json:"step,omitempty"` // "node" 或 "clawhub"
	Message string `json:"message,omitempty"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path,omitempty"`
}
