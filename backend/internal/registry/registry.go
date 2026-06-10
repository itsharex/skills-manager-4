package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/internal/version"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// SkillRegistryEntry 注册表中单个技能条目（扩展版）
type SkillRegistryEntry struct {
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	Source        models.Source           `json:"source"`
	Versions      map[string]models.SkillVersion `json:"versions"`
	LatestVersion string                  `json:"latest_version"`
	UserTags      []string                `json:"user_tags,omitempty"` // 用户自定义标签
}

// Registry 管理 skills-registry.json
type Registry struct {
	path string
	data *models.Registry
}

// Load 从 skillspool 根目录加载注册表
func Load(skillspoolRoot string) (*Registry, error) {
	path := filepath.Join(skillspoolRoot, "skills-registry.json")
	r := &Registry{
		path: path,
		data: &models.Registry{
			Version:     1,
			InstalledAt: time.Now().UTC().Format(time.RFC3339),
			Skills:      make(map[string]*models.Skill),
		},
	}

	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, r.data); err != nil {
			return nil, fmt.Errorf("parse registry: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read registry: %w", err)
	}

	return r, nil
}

// Save 保存注册表
func (r *Registry) Save() error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, 0o644)
}

// AddSkill 添加或更新技能
func (r *Registry) AddSkill(skillName string, info *models.SkillInfo, source models.Source, versionStr, relPath string, agentIDs []string) {
	now := time.Now().UTC().Format(time.RFC3339)

	skill, ok := r.data.Skills[skillName]
	if !ok {
		skill = &models.Skill{
			Name:          skillName,
			Description:   info.Description,
			Source:        source,
			Versions:      make(map[string]models.SkillVersion),
			LatestVersion: versionStr,
		}
		r.data.Skills[skillName] = skill
	}

	// 更新描述
	if info.Description != "" {
		skill.Description = info.Description
	}

	// 添加或更新版本
	sv := models.SkillVersion{
		Version:     versionStr,
		InstalledAt: now,
		Path:        relPath,
		Agents:      uniqueStrings(append(svAgents(skill, versionStr), agentIDs...)),
	}
	skill.Versions[versionStr] = sv

	// 更新 latest
	allVersions := make([]string, 0, len(skill.Versions))
	for v := range skill.Versions {
		allVersions = append(allVersions, v)
	}
	skill.LatestVersion = version.Latest(allVersions)
}

// List 返回所有已安装技能（按名称排序）
func (r *Registry) List() []*models.Skill {
	names := make([]string, 0, len(r.data.Skills))
	for name := range r.data.Skills {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]*models.Skill, len(names))
	for i, name := range names {
		s := r.data.Skills[name]
		// 深拷贝避免外部修改
		skill := *s
		skill.Versions = make(map[string]models.SkillVersion)
		for k, v := range s.Versions {
			skill.Versions[k] = v
		}
		result[i] = &skill
	}
	return result
}

// --- 标签管理 ---

// AddUserTag 给指定技能添加一个用户自定义标签
func (r *Registry) AddUserTag(skillName, tag string) bool {
	s, ok := r.data.Skills[skillName]
	if !ok {
		return false
	}
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return false
	}
	for _, t := range s.UserTags {
		if t == tag {
			return false
		}
	}
	s.UserTags = append(s.UserTags, tag)
	_ = r.Save()
	return true
}

// RemoveUserTag 从指定技能移除一个标签
func (r *Registry) RemoveUserTag(skillName, tag string) bool {
	s, ok := r.data.Skills[skillName]
	if !ok {
		return false
	}
	newTags := make([]string, 0, len(s.UserTags))
	for _, t := range s.UserTags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	if len(newTags) == len(s.UserTags) {
		return false
	}
	s.UserTags = newTags
	_ = r.Save()
	return true
}

// GetUserTags 获取指定技能的用户标签
func (r *Registry) GetUserTags(skillName string) []string {
	if s, ok := r.data.Skills[skillName]; ok {
		result := make([]string, len(s.UserTags))
		copy(result, s.UserTags)
		return result
	}
	return nil
}

// GetAllTags 返回所有技能的标签使用情况
func (r *Registry) GetAllTags() []models.TagUsage {
	counts := make(map[string]int)
	for _, s := range r.data.Skills {
		for _, t := range s.UserTags {
			counts[t]++
		}
	}
	var result []models.TagUsage
	for tag, cnt := range counts {
		result = append(result, models.TagUsage{Tag: tag, Count: cnt})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Tag < result[j].Tag
		}
		return result[i].Count > result[j].Count
	})
	return result
}

// --- 辅助函数 ---

// CollectAgents 返回技能安装到的 Agent ID 集合（所有版本去重后）
func (r *Registry) CollectAgents(skillName string) []string {
	s, ok := r.data.Skills[skillName]
	if !ok {
		return nil
	}
	seen := make(map[string]bool)
	var result []string
	for _, v := range s.Versions {
		for _, a := range v.Agents {
			if !seen[a] {
				seen[a] = true
				result = append(result, a)
			}
		}
	}
	sort.Strings(result)
	return result
}

// AddAgentToVersion 为指定技能版本添加 Agent 安装记录
func (r *Registry) AddAgentToVersion(skillName, version, agentID string) bool {
	s, ok := r.data.Skills[skillName]
	if !ok {
		return false
	}
	v, ok := s.Versions[version]
	if !ok {
		// 使用 latest 版本作为回退
		if s.LatestVersion != "" {
			v = s.Versions[s.LatestVersion]
			version = s.LatestVersion
		}
		if v.Version == "" {
			return false
		}
	}
	for _, a := range v.Agents {
		if a == agentID {
			return false
		}
	}
	v.Agents = append(v.Agents, agentID)
	sort.Strings(v.Agents)
	s.Versions[version] = v
	_ = r.Save()
	return true
}

// RemoveAgentFromVersion 从指定技能版本移除 Agent 记录
func (r *Registry) RemoveAgentFromVersion(skillName, version, agentID string) bool {
	s, ok := r.data.Skills[skillName]
	if !ok {
		return false
	}
	v, ok := s.Versions[version]
	if !ok {
		return false
	}
	newAgents := make([]string, 0, len(v.Agents))
	for _, a := range v.Agents {
		if a != agentID {
			newAgents = append(newAgents, a)
		}
	}
	if len(newAgents) == len(v.Agents) {
		return false
	}
	v.Agents = newAgents
	s.Versions[version] = v
	_ = r.Save()
	return true
}

// AddSkillDirect 直接添加/更新技能（用在迁移操作等场景）
func (r *Registry) AddSkillDirect(skillName string, info *models.SkillInfo, source models.Source, versionStr, relPath string, agentIDs []string) {
	r.AddSkill(skillName, info, source, versionStr, relPath, agentIDs)
}

// Get 获取指定技能
func (r *Registry) Get(name string) (*models.Skill, bool) {
	s, ok := r.data.Skills[name]
	return s, ok
}

// Remove 删除技能（及其某版本）
func (r *Registry) Remove(name, versionStr string) {
	if s, ok := r.data.Skills[name]; ok {
		if versionStr == "" {
			delete(r.data.Skills, name)
			return
		}
		delete(s.Versions, versionStr)
		if len(s.Versions) == 0 {
			delete(r.data.Skills, name)
			return
		}
		// 重新计算 latest
		all := make([]string, 0, len(s.Versions))
		for v := range s.Versions {
			all = append(all, v)
		}
		s.LatestVersion = version.Latest(all)
	}
}

// --- 工具 ---

func svAgents(s *models.Skill, v string) []string {
	if sv, ok := s.Versions[v]; ok {
		return sv.Agents
	}
	return nil
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
