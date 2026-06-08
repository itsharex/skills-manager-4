package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	skillsapi "github.com/skillsmanager/skillsmanager/backend/pkg/api"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// App 是 Wails 应用的主结构体
type App struct {
	api *skillsapi.API
}

// NewApp 创建新的 App
func NewApp() *App {
	return &App{}
}

// OnStartup 在应用启动时调用：初始化 API
func (a *App) OnStartup(ctx context.Context) {
	// 配置路径为空时使用系统默认路径
	var err error
	a.api, err = skillsapi.New("")
	if err != nil {
		runtime.LogError(ctx, fmt.Sprintf("startup: init api: %v", err))
	}
}

// OnShutdown 在应用关闭时调用
func (a *App) OnShutdown(ctx context.Context) {
}

// ---------------- Frontend bound methods ----------------

// GetConfig 返回当前配置（skillspool 根目录 + Agent 列表
func (a *App) GetConfig() *models.Config {
	if a.api == nil {
		return nil
	}
	return a.api.Config()
}

// SkillspoolRoot 返回技能存储的根目录
func (a *App) SkillspoolRoot() string {
	if a.api == nil {
		return ""
	}
	return a.api.SkillspoolRoot()
}

// ListSkills 返回所有已安装的技能
func (a *App) ListSkills() []*models.Skill {
	if a.api == nil {
		return nil
	}
	return a.api.ListSkills()
}

// ListAgents 返回所有 Agent 配置
func (a *App) ListAgents() map[string]models.Agent {
	if a.api == nil {
		return nil
	}
	return a.api.ListAgents()
}

// Install 安装技能到选定的 Agent
func (a *App) Install(req skillsapi.InstallRequest) ([]skillsapi.InstallResult, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.Install(req)
}

// --- 新增：技能市场相关 API

// GetMarketConfig 获取技能市场配置
func (a *App) GetMarketConfig() *models.SkillMarketConfig {
	if a.api == nil || a.api.Config() == nil {
		return nil
	}
	cfg := a.api.Config()
	return &cfg.SkillMarket
}

// SetMarketConfig 设置技能市场配置
func (a *App) SetMarketConfig(config models.SkillMarketConfig) error {
	if a.api == nil {
		return fmt.Errorf("应用尚未初始化")
	}
	return a.api.SetMarketConfig(config)
}

// ScanMarket 扫描技能市场
func (a *App) ScanMarket() (*models.ScanMarketResult, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.ScanMarket()
}

// ListMarketSkills 按分类列出市场技能
func (a *App) ListMarketSkills(category string) ([]models.MarketSkill, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.ListMarketSkills(category)
}

// --- 新增：全局技能库扫描

// ListGlobalSkillsWithAgents 列出所有 Agent 的全局技能库
func (a *App) ListGlobalSkillsWithAgents() ([]models.GlobalSkillWithAgents, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.ListGlobalSkillsWithAgents()
}

// --- 新增：项目技能管理

// ScanProjectSkills 扫描项目中的技能
func (a *App) ScanProjectSkills(projectPath string) ([]models.ProjectSkill, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.ScanProjectSkills(projectPath)
}

// MigrateProjectSkillToLibrary 迁移项目技能到库
func (a *App) MigrateProjectSkillToLibrary(skillPath string, projectPath string) (*models.MigrateResult, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.MigrateProjectSkillToLibrary(skillPath, projectPath)
}

// --- 新增：批量同步

// BatchSyncSkills 批量同步技能
func (a *App) BatchSyncSkills(req models.BatchSyncRequest) (*models.BatchSyncResult, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.BatchSyncSkills(req)
}

// --- 清理功能 API ---

// UninstallFromProject 从项目中移除技能
func (a *App) UninstallFromProject(skillName, agentID, projectPath string) error {
	if a.api == nil {
		return fmt.Errorf("应用尚未初始化")
	}
	return a.api.UninstallFromProject(skillName, agentID, projectPath)
}

// CleanOrphanedSymlinks 清理孤立软链
func (a *App) CleanOrphanedSymlinks(agentID string) (int, error) {
	if a.api == nil {
		return 0, fmt.Errorf("应用尚未初始化")
	}
	return a.api.CleanOrphanedSymlinks(agentID)
}

// CleanGlobalLibrary 清理全局库
func (a *App) CleanGlobalLibrary(dryRun bool) (*models.CleanResult, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.CleanGlobalLibrary(dryRun)
}

// BatchClean 批量清理
func (a *App) BatchClean(criteria models.BatchCleanCriteria) (*models.BatchCleanResult, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.BatchClean(criteria)
}

// --- 版本管理 API ---

// ListSkillVersions 列出技能版本
func (a *App) ListSkillVersions(skillName string) ([]models.VersionInfo, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.ListSkillVersions(skillName)
}

// SwitchSkillVersion 切换技能版本
func (a *App) SwitchSkillVersion(skillName, version string) error {
	if a.api == nil {
		return fmt.Errorf("应用尚未初始化")
	}
	return a.api.SwitchSkillVersion(skillName, version)
}

// DeleteSkillVersion 删除技能版本
func (a *App) DeleteSkillVersion(skillName, version string) error {
	if a.api == nil {
		return fmt.Errorf("应用尚未初始化")
	}
	return a.api.DeleteSkillVersion(skillName, version)
}

// --- 健康检查 API ---

// CheckHealth 执行健康检查
func (a *App) CheckHealth() (*models.HealthReport, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.CheckHealth()
}

// FixBrokenSymlinks 修复孤立软链
func (a *App) FixBrokenSymlinks(agentID string) (int, error) {
	if a.api == nil {
		return 0, fmt.Errorf("应用尚未初始化")
	}
	return a.api.FixBrokenSymlinks(agentID)
}

// --- 统计 API ---

// GetSkillStats 获取技能统计
func (a *App) GetSkillStats(skillName string) (*models.SkillStats, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.GetSkillStats(skillName)
}

// GetAgentStats 获取 Agent 统计
func (a *App) GetAgentStats(agentID string) (*models.AgentStats, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.GetAgentStats(agentID)
}

// GetUsageDashboard 获取使用统计仪表盘
func (a *App) GetUsageDashboard() (*models.UsageDashboard, error) {
	if a.api == nil {
		return nil, fmt.Errorf("应用尚未初始化")
	}
	return a.api.GetUsageDashboard()
}

// 为 Wails 生成绑定代码时引用这些类型（避免编译器警告/未使用
var (
	_ = models.Skill{}
	_ = models.Agent{}
	_ = models.Source{}
	_ = models.SkillVersion{}
	_ = models.Config{}
	_ = models.SkillspoolConfig{}
	_ = models.SkillMarketConfig{}
	_ = models.MarketSkill{}
	_ = models.ScanMarketResult{}
	_ = models.GlobalSkillWithAgents{}
	_ = models.AgentSkillStatus{}
	_ = models.ProjectSkill{}
	_ = models.MigrateResult{}
	_ = models.BatchSyncRequest{}
	_ = models.BatchSyncResult{}
	_ = models.BatchSyncItemResult{}
	_ = models.CleanResult{}
	_ = models.CleanItemResult{}
	_ = models.BatchCleanCriteria{}
	_ = models.BatchCleanResult{}
	_ = models.VersionInfo{}
	_ = models.VersionCompare{}
	_ = models.VersionDiff{}
	_ = models.HealthReport{}
	_ = models.HealthSummary{}
	_ = models.HealthIssue{}
	_ = models.SymlinkIssue{}
	_ = models.FileIssue{}
	_ = models.VersionIssue{}
	_ = models.SkillStats{}
	_ = models.AgentInstall{}
	_ = models.AgentStats{}
	_ = models.SkillInstall{}
	_ = models.UsageDashboard{}
	_ = models.SkillCount{}
	_ = models.SkillActivity{}
	_ = models.ActivityEntry{}
	_ = skillsapi.InstallRequest{}
	_ = skillsapi.InstallResult{}
	_ = skillsapi.InstallLink{}
)
