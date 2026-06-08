package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/internal/version"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

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
