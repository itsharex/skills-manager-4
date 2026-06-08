package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Storage 管理 skillspool 目录
type Storage struct {
	root string
}

// New 创建 Storage，root 为 skillspool 根目录（可含 ~）
func New(root string) (*Storage, error) {
	root = expandPath(root)
	if root == "" {
		root = defaultRoot()
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create skillspool root: %w", err)
	}
	return &Storage{root: root}, nil
}

// Root 返回 skillspool 根目录
func (s *Storage) Root() string {
	return s.root
}

// SkillPath 返回指定技能的目录
// 支持 @scope/name 格式自动拆分:
//   "frontend-design"    → {root}/frontend-design
//   "@myorg/my-skill"    → {root}/@myorg/my-skill
//   "patterns/category"  → {root}/patterns/category
func (s *Storage) SkillPath(skillName string) string {
	parts := splitAtPath(skillName)
	parts = append([]string{s.root}, parts...)
	return filepath.Join(parts...)
}

// VersionPath 返回指定技能版本的目录
//   skillspool/<skill>/<version>
func (s *Storage) VersionPath(skillName, version string) string {
	return filepath.Join(s.SkillPath(skillName), version)
}

// LatestPath 返回 latest 软连接路径
func (s *Storage) LatestPath(skillName string) string {
	return filepath.Join(s.SkillPath(skillName), "latest")
}

// EnsureVersionDir 创建技能版本目录
func (s *Storage) EnsureVersionDir(skillName, version string) (string, error) {
	p := s.VersionPath(skillName, version)
	if err := os.MkdirAll(p, 0o755); err != nil {
		return "", fmt.Errorf("create version dir: %w", err)
	}
	return p, nil
}

// ListVersions 列出技能的所有已安装版本
func (s *Storage) ListVersions(skillName string) ([]string, error) {
	base := s.SkillPath(skillName)
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var versions []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "latest" {
			continue
		}
		versions = append(versions, name)
	}
	return versions, nil
}

// UpdateLatest 更新 latest 软连接到指定版本
func (s *Storage) UpdateLatest(skillName, version string) error {
	base := s.SkillPath(skillName)
	versionDir := filepath.Join(base, version)
	if _, err := os.Stat(versionDir); err != nil {
		return fmt.Errorf("version dir not exist: %w", err)
	}

	latestPath := filepath.Join(base, "latest")

	// 如果已存在，先删除（可能是旧的软连接或目录）
	if info, err := os.Lstat(latestPath); err == nil {
		// 如果已经是正确的软连接，跳过
		if info.Mode()&os.ModeSymlink != 0 {
			if target, err := os.Readlink(latestPath); err == nil {
				// 目标可能是相对或绝对路径
				if target == version || target == versionDir {
					return nil
				}
			}
		}
		if err := os.RemoveAll(latestPath); err != nil {
			return fmt.Errorf("remove old latest: %w", err)
		}
	}

	// 创建软连接
	if err := os.Symlink(version, latestPath); err != nil {
		// Windows 可能需要特殊处理，这里降级为目录复制
		if runtime.GOOS == "windows" {
			return copyDir(versionDir, latestPath)
		}
		return fmt.Errorf("create latest symlink: %w", err)
	}
	return nil
}

// --- 工具函数 ---

// splitAtPath 将技能名按路径分隔符拆分，且保留 @scope
//   "frontend-design"         → ["frontend-design"]
//   "@myorg/my-skill"         → ["@myorg", "my-skill"]
//   "patterns/category/name"  → ["patterns", "category", "name"]
func splitAtPath(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return []string{"_unknown_"}
	}

	// 统一路径分隔符为 /
	cleaned := filepath.ToSlash(name)
	parts := strings.Split(cleaned, "/")

	// 过滤空串
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, sanitizeName(p))
		}
	}
	if len(result) == 0 {
		return []string{"_unknown_"}
	}
	return result
}

// sanitizeName 清理文件名，去除路径不安全字符
func sanitizeName(s string) string {
	replacer := strings.NewReplacer(
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "-",
	)
	return replacer.Replace(s)
}

// expandPath 展开 ~ 和环境变量
func expandPath(p string) string {
	p = os.ExpandEnv(p)
	if strings.HasPrefix(p, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

// defaultRoot 根据操作系统返回默认 skillspool 根目录
func defaultRoot() string {
	switch runtime.GOOS {
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", "SkillsManager", "skillspool")
		}
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "SkillsManager", "skillspool")
		}
	default:
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".local", "share", "skillsmanager", "skillspool")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".skillsmanager", "skillspool")
}

// copyDir 递归复制目录（用于软连接不可用时的降级策略）
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
