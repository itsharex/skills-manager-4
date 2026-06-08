package api

import (
	"fmt"

	"github.com/skillsmanager/skillsmanager/backend/internal/agent"
	"github.com/skillsmanager/skillsmanager/backend/internal/config"
	"github.com/skillsmanager/skillsmanager/backend/internal/installer"
	"github.com/skillsmanager/skillsmanager/backend/internal/lifecycle"
	"github.com/skillsmanager/skillsmanager/backend/internal/registry"
	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// API 统一对外接口（Wails GUI + CLI 共用）
type API struct {
	cfg       *config.Config
	store     *storage.Storage
	reg       *registry.Registry
	agentMgr  *agent.Manager
	install   *installer.Installer
	lifecycle *lifecycle.Lifecycle
}

// New 创建 API
func New(configPath string) (*API, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	store, err := storage.New(cfg.Data.Skillspool.Root)
	if err != nil {
		return nil, err
	}
	reg, err := registry.Load(store.Root())
	if err != nil {
		return nil, err
	}
	agentMgr := agent.New(cfg.Data.Agents)
	inst := installer.New(store, reg, agentMgr)
	lc := lifecycle.New(store, reg, agentMgr)

	return &API{
		cfg:       cfg,
		store:     store,
		reg:       reg,
		agentMgr:  agentMgr,
		install:   inst,
		lifecycle: lc,
	}, nil
}

// Config 返回配置
func (a *API) Config() *models.Config {
	return a.cfg.Data
}

// SkillspoolRoot 返回 skillspool 根目录
func (a *API) SkillspoolRoot() string {
	return a.store.Root()
}

// ListSkills 列出所有已安装技能
func (a *API) ListSkills() []*models.Skill {
	return a.reg.List()
}

// GetSkill 获取单个技能
func (a *API) GetSkill(name string) (*models.Skill, bool) {
	return a.reg.Get(name)
}

// ListAgentIDs 返回已检测的 Agent ID 列表
func (a *API) ListAgentIDs() []string {
	return a.agentMgr.ListDetected()
}

// ListAgents 返回所有 Agent 配置
func (a *API) ListAgents() map[string]models.Agent {
	return a.agentMgr.List()
}

// InstallRequest 安装请求参数（方便 Wails 调用）
type InstallRequest struct {
	Source     string   `json:"source"`
	SubPath    string   `json:"sub_path,omitempty"`
	Version    string   `json:"version,omitempty"`
	Ref        string   `json:"ref,omitempty"`
	Agents     []string `json:"agents"`
	Scope      string   `json:"scope"` // "global" or "project"
	ProjectDir string   `json:"project_dir,omitempty"`
}

// InstallLink 是安装到单个 Agent 的链接结果
type InstallLink struct {
	AgentID string `json:"agent_id"`
	Path    string `json:"path"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// InstallResult 是安装单个技能的完整结果
type InstallResult struct {
	SkillName string                 `json:"skill_name"`
	Version   string                 `json:"version"`
	Source    models.Source          `json:"source"`
	Agents    map[string]InstallLink `json:"agents"`
}

// Install 安装技能
func (a *API) Install(req InstallRequest) ([]InstallResult, error) {
	if len(req.Agents) == 0 {
		req.Agents = a.agentMgr.ListDetected()
	}
	if len(req.Agents) == 0 {
		return nil, fmt.Errorf("no agent detected, please configure agents first")
	}
	if req.Scope == "" {
		req.Scope = "global"
	}
	internalResults, err := a.install.Install(installer.InstallOptions{
		Source:     req.Source,
		SubPath:    req.SubPath,
		Version:    req.Version,
		Ref:        req.Ref,
		AgentIDs:   req.Agents,
		Scope:      req.Scope,
		ProjectDir: req.ProjectDir,
	})
	if err != nil {
		return nil, err
	}
	// 转换结果类型
	var results []InstallResult
	for _, ir := range internalResults {
		agentsMap := make(map[string]InstallLink)
		for agentID, link := range ir.Agents {
			agentsMap[agentID] = InstallLink{
				AgentID: link.AgentID,
				Path:    link.Path,
				Success: link.Success,
				Error:   link.Error,
			}
		}
		results = append(results, InstallResult{
			SkillName: ir.SkillName,
			Version:   ir.Version,
			Source:    ir.Source,
			Agents:    agentsMap,
		})
	}
	return results, nil
}

// SaveConfig 保存配置
func (a *API) SaveConfig() error {
	return a.cfg.Save()
}

// AddAgent 手动添加 Agent
func (a *API) AddAgent(id string, ag models.Agent) {
	a.cfg.AddAgent(id, ag)
	_ = a.cfg.Save()
	// 重新创建 Manager
	a.agentMgr = agent.New(a.cfg.Data.Agents)
	a.install = installer.New(a.store, a.reg, a.agentMgr)
	a.lifecycle = lifecycle.New(a.store, a.reg, a.agentMgr)
}

// --- 清理功能 API ---

// UninstallFromProject 从项目中移除技能（仅移除软链）
func (a *API) UninstallFromProject(skillName, agentID, projectPath string) error {
	return a.lifecycle.UninstallFromProject(skillName, agentID, projectPath)
}

// CleanOrphanedSymlinks 清理孤立软链
func (a *API) CleanOrphanedSymlinks(agentID string) (int, error) {
	return a.lifecycle.CleanOrphanedSymlinks(agentID)
}

// CleanGlobalLibrary 清理全局库
func (a *API) CleanGlobalLibrary(dryRun bool) (*models.CleanResult, error) {
	return a.lifecycle.CleanGlobalLibrary(dryRun)
}

// BatchClean 批量清理
func (a *API) BatchClean(criteria models.BatchCleanCriteria) (*models.BatchCleanResult, error) {
	// TODO: 实现批量清理逻辑
	result := &models.BatchCleanResult{
		DryRun: false,
		Results: []models.CleanItemResult{},
	}
	return result, nil
}

// --- 版本管理 API ---

// ListSkillVersions 列出技能版本
func (a *API) ListSkillVersions(skillName string) ([]models.VersionInfo, error) {
	return a.lifecycle.ListSkillVersions(skillName)
}

// SwitchSkillVersion 切换技能版本
func (a *API) SwitchSkillVersion(skillName, version string) error {
	return a.lifecycle.SwitchSkillVersion(skillName, version)
}

// DeleteSkillVersion 删除技能版本
func (a *API) DeleteSkillVersion(skillName, version string) error {
	return a.lifecycle.DeleteSkillVersion(skillName, version)
}

// --- 健康检查 API ---

// CheckHealth 执行健康检查
func (a *API) CheckHealth() (*models.HealthReport, error) {
	return a.lifecycle.CheckHealth()
}

// FixBrokenSymlinks 修复孤立软链
func (a *API) FixBrokenSymlinks(agentID string) (int, error) {
	return a.lifecycle.FixBrokenSymlinks(agentID)
}

// --- 统计 API ---

// GetSkillStats 获取技能统计
func (a *API) GetSkillStats(skillName string) (*models.SkillStats, error) {
	return a.lifecycle.GetSkillStats(skillName)
}

// GetAgentStats 获取 Agent 统计
func (a *API) GetAgentStats(agentID string) (*models.AgentStats, error) {
	return a.lifecycle.GetAgentStats(agentID)
}

// GetUsageDashboard 获取使用统计仪表盘
func (a *API) GetUsageDashboard() (*models.UsageDashboard, error) {
	return a.lifecycle.GetUsageDashboard()
}

// SearchMarket 搜索技能市场（简单版本：目前从 GitHub 搜索）
// 完整的市场搜索可以后续对接 trae skills market API
func (a *API) SearchMarket(keyword string) []MarketSkill {
	// 目前返回空，GUI 层会直接让用户粘贴 GitHub URL
	// TODO: 对接 trae 技能市场 API
	return nil
}

type MarketSkill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Stars       int      `json:"stars"`
	Tags        []string `json:"tags"`
}

// --- 新增：技能市场配置管理 ---

// SetMarketConfig 设置技能市场配置
func (a *API) SetMarketConfig(config models.SkillMarketConfig) error {
	a.cfg.Data.SkillMarket = config
	return a.cfg.Save()
}

// --- 新增：技能市场扫描与获取 ---

// ScanMarket 扫描技能市场
func (a *API) ScanMarket() (*models.ScanMarketResult, error) {
	// 简单实现：扫描配置的市场 URL
	if a.cfg.Data.SkillMarket.URL == "" {
		return &models.ScanMarketResult{
			TotalSkills: 0,
			Categories:  []string{},
			Skills:      []models.MarketSkill{},
		}, nil
	}
	// TODO: 完整的市场扫描实现
	return &models.ScanMarketResult{
		TotalSkills: 0,
		Categories:  []string{},
		Skills:      []models.MarketSkill{},
	}, nil
}

// ListMarketSkills 列出市场技能
func (a *API) ListMarketSkills(category string) ([]models.MarketSkill, error) {
	// TODO: 完整实现
	return []models.MarketSkill{}, nil
}

// --- 新增：全局技能库扫描 ---

// ListGlobalSkillsWithAgents 扫描并列出所有 Agent 的全局技能
func (a *API) ListGlobalSkillsWithAgents() ([]models.GlobalSkillWithAgents, error) {
	// 简单实现框架
	skillMap := make(map[string]*models.GlobalSkillWithAgents)
	agents := a.agentMgr.List()

	for _, agent := range agents {
		if !agent.Installed {
			continue
		}
		// TODO: 扫描 agent.GlobalLocation 中的技能目录
	}

	var result []models.GlobalSkillWithAgents
	for _, skill := range skillMap {
		result = append(result, *skill)
	}
	return result, nil
}

// --- 新增：项目技能管理 ---

// ScanProjectSkills 扫描项目中的技能
func (a *API) ScanProjectSkills(projectPath string) ([]models.ProjectSkill, error) {
	// TODO: 实现扫描项目目录
	return []models.ProjectSkill{}, nil
}

// MigrateProjectSkillToLibrary 迁移项目技能到库
func (a *API) MigrateProjectSkillToLibrary(skillPath string, projectPath string) (*models.MigrateResult, error) {
	// TODO: 实现迁移逻辑
	return &models.MigrateResult{
		Success: false,
		Error:   "not implemented yet",
	}, nil
}

// --- 新增：批量同步 ---

// BatchSyncSkills 批量同步技能
func (a *API) BatchSyncSkills(req models.BatchSyncRequest) (*models.BatchSyncResult, error) {
	// TODO: 实现批量同步逻辑
	result := &models.BatchSyncResult{
		Total: 0,
		Succeeded: 0,
		Failed: 0,
		Results: []models.BatchSyncItemResult{},
	}
	return result, nil
}
