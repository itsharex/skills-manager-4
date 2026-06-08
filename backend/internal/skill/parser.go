package skill

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
	"gopkg.in/yaml.v3"
)

// ParseSkillFile 解析 SKILL.md 文件
// 支持两种格式:
// 1. YAML frontmatter: 以 --- 开头和结束的 YAML 块
// 2. 简单 Markdown: 第一个 # 标题作为 name，第一段作为 description
func ParseSkillFile(path string) (*models.SkillInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read SKILL.md: %w", err)
	}

	return parseSkillContent(string(data))
}

// ScanDirectory 扫描目录，查找所有包含 SKILL.md 的子目录
// 返回每个技能的 路径 + 解析后的元信息
func ScanDirectory(root string) (map[string]*models.SkillInfo, error) {
	results := make(map[string]*models.SkillInfo)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// 跳过隐藏目录和 .git
			name := d.Name()
			if path != root && (strings.HasPrefix(name, ".") || name == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "SKILL.md" {
			return nil
		}

		dir := filepath.Dir(path)
		info, err := ParseSkillFile(path)
		if err != nil {
			// 解析失败的跳过，但记录错误
			return nil
		}
		// 如果 name 没从 frontmatter 解析到，使用目录名
		if info.Name == "" {
			info.Name = filepath.Base(dir)
		}
		results[dir] = info
		return nil
	})

	if err != nil {
		return nil, err
	}
	return results, nil
}

// --- 内部解析 ---

type frontmatter struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Version     string   `yaml:"version"`
	Author      string   `yaml:"author"`
	Tags        []string `yaml:"tags"`
}

func parseSkillContent(content string) (*models.SkillInfo, error) {
	content = strings.TrimSpace(content)

	// 尝试解析 YAML frontmatter
	if strings.HasPrefix(content, "---") {
		// 找到结束标记
		after := strings.TrimPrefix(content, "---")
		after = strings.TrimLeft(after, "\r\n")

		endIdx := strings.Index(after, "---")
		if endIdx > 0 {
			fmText := strings.TrimSpace(after[:endIdx])
			var fm frontmatter
			if err := yaml.Unmarshal([]byte(fmText), &fm); err == nil {
				return &models.SkillInfo{
					Name:        strings.TrimSpace(fm.Name),
					Description: strings.TrimSpace(fm.Description),
					Version:     strings.TrimSpace(fm.Version),
					Author:      strings.TrimSpace(fm.Author),
					Tags:        fm.Tags,
				}, nil
			}
		}
	}

	// 回退: 从 Markdown 中提取
	return parseSimpleMarkdown(content), nil
}

func parseSimpleMarkdown(content string) *models.SkillInfo {
	info := &models.SkillInfo{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	foundHeading := false
	var descLines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !foundHeading && strings.HasPrefix(line, "# ") {
			info.Name = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			foundHeading = true
			continue
		}
		if foundHeading && !strings.HasPrefix(line, "#") {
			descLines = append(descLines, line)
			if len(descLines) >= 3 {
				break
			}
		}
	}

	info.Description = strings.Join(descLines, " ")
	if len(info.Description) > 200 {
		info.Description = info.Description[:200] + "..."
	}
	return info
}
