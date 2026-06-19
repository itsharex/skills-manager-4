// Package clawhub 封装 ClawHub 技能市场（OpenClaw 官方技能市场）的搜索/下载能力。
//
// 协议说明：
//   - ClawHub 没有公开的 HTTP 搜索 API；其内容托管在 GitHub 仓库 `openclaw-dev/skills`。
//   - 仓库结构：仓库根下若干子目录（owner），owner 下若干子目录（slug），slug 下含 SKILL.md。
//   - 实现策略：
//       1) 通过 GitHub Trees API 列出仓库内全部 SKILL.md 文件
//       2) 搜索时按 owner/slug 名称 + SKILL.md 头部的 name/description 字段做关键词匹配
//       3) 安装时直接通过 raw.githubusercontent.com 拉取 owner/slug 目录内容
//   - 同时保留对 `clawhub` CLI 的兼容（如果用户机器上已安装）。
package clawhub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// 默认上游仓库（可通过 SetRegistry 修改）
const defaultRegistry = "openclaw-dev/skills"

// Manager 封装 ClawHub 市场能力。
type Manager struct {
	skillspoolRoot string
	registry       string
	cache          *registryCache
	cliInstalled   bool
	cliMu          sync.Mutex
	cliChecked     bool
	httpClient     *http.Client
}

// registryCache 缓存已扫描到的 skill 列表
type registryCache struct {
	mu        sync.RWMutex
	loaded    bool
	skills    []models.ClawHubSkill
	loadedAt  time.Time
	cacheTTL  time.Duration
}

// New 创建一个 Manager
func New(skillspoolRoot string) *Manager {
	return &Manager{
		skillspoolRoot: skillspoolRoot,
		registry:       defaultRegistry,
		cache:          &registryCache{cacheTTL: 10 * time.Minute},
		httpClient:     &http.Client{Timeout: 20 * time.Second},
	}
}

// SetRegistry 允许切换为镜像或其它 fork
func (m *Manager) SetRegistry(repo string) {
	m.registry = strings.TrimSpace(repo)
	if m.registry == "" {
		m.registry = defaultRegistry
	}
	m.cache.mu.Lock()
	m.cache.loaded = false
	m.cache.skills = nil
	m.cache.mu.Unlock()
}

// HasCLI 检测本机是否安装了 clawhub CLI
func (m *Manager) HasCLI() bool {
	m.cliMu.Lock()
	defer m.cliMu.Unlock()
	if m.cliChecked {
		return m.cliInstalled
	}
	m.cliChecked = true
	if _, err := exec.LookPath("clawhub"); err == nil {
		m.cliInstalled = true
		return true
	}
	m.cliInstalled = false
	return false
}

// ---------- 运行时状态 ----------

// RuntimeStatus 检测运行时环境
func (m *Manager) RuntimeStatus() models.RuntimeStatus {
	st := models.RuntimeStatus{}
	if nodePath, err := exec.LookPath("node"); err == nil {
		st.NodeInstalled = true
		st.NodePath = nodePath
		if out, err := exec.Command(nodePath, "--version").Output(); err == nil {
			st.NodeVersion = strings.TrimSpace(string(out))
		}
	}
	if _, err := exec.LookPath("npm"); err == nil {
		st.HasNpm = true
	}
	st.ClawHubInstalled = m.HasCLI()
	if st.ClawHubInstalled {
		if p, err := exec.LookPath("clawhub"); err == nil {
			st.ClawHubPath = p
			if out, err := exec.Command(p, "--version").Output(); err == nil {
				st.ClawHubVersion = strings.TrimSpace(string(out))
			}
		}
	}
	// 探测网络可达的 registry
	if st.ClawHubInstalled || st.NodeInstalled {
		st.RegistryReachable = m.pingRegistry()
		st.RegistryName = m.registry
	}
	if !st.NodeInstalled {
		st.Message = "未检测到 Node.js（不影响 GitHub 方式浏览，但建议安装以启用 clawhub CLI）。"
	} else if st.ClawHubInstalled {
		if !st.RegistryReachable {
			st.Message = "已安装 clawhub CLI，但市场网络不通，将降级为 GitHub 源。"
		} else {
			st.Message = "已安装 clawhub CLI，可直接搜索/安装技能。"
		}
	} else {
		st.Message = "未安装 clawhub CLI，将通过 GitHub Registry 拉取技能（无需 CLI 即可使用）。"
	}
	return st
}

// EnsureRuntime 保留兼容：尝试安装 CLI（如已存在则跳过）
func (m *Manager) EnsureRuntime() (*models.RuntimeStatus, error) {
	st := m.RuntimeStatus()
	if st.ClawHubInstalled {
		return &st, nil
	}
	if !st.NodeInstalled || !st.HasNpm {
		// 允许"仅 GitHub 模式"运行
		return &st, nil
	}
	cmd := exec.Command("npm", "install", "-g", "clawhub@latest")
	var buf strings.Builder
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		// 安装失败也不阻塞，GitHub 模式仍可用
		return &st, nil
	}
	// 重新检测
	m.cliMu.Lock()
	m.cliChecked = false
	m.cliMu.Unlock()
	newSt := m.RuntimeStatus()
	return &newSt, nil
}

// pingRegistry 测试能否访问 GitHub 上的 registry 仓库
func (m *Manager) pingRegistry() bool {
	url := fmt.Sprintf("https://api.github.com/repos/%s", m.registry)
	req, _ := http.NewRequest("HEAD", url, nil)
	req.Header.Set("User-Agent", "skills-manager")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// ---------- 搜索 ----------

// Search 返回与关键词匹配的技能列表
// 优先用 GitHub API 扫描 registry；如已安装 CLI 且 keyword 为空则 fallback 到 CLI。
func (m *Manager) Search(keyword string) ([]models.ClawHubSkill, error) {
	// 1. 加载 registry 缓存
	all, err := m.loadRegistrySkills()
	if err != nil {
		// 网络失败时尝试 CLI 兜底
		if m.HasCLI() {
			return m.searchViaCLI(keyword)
		}
		return nil, fmt.Errorf("加载 ClawHub 列表失败: %w", err)
	}

	// 2. 关键词过滤
	kw := strings.ToLower(strings.TrimSpace(keyword))
	if kw == "" {
		return all, nil
	}
	var matched []models.ClawHubSkill
	for _, s := range all {
		if matchSkill(s, kw) {
			matched = append(matched, s)
		}
	}
	// 按相关度排序：owner/slug 完全匹配 > 名称匹配 > 描述匹配
	sort.SliceStable(matched, func(i, j int) bool {
		return relevanceScore(matched[i], kw) > relevanceScore(matched[j], kw)
	})
	return matched, nil
}

// searchViaCLI 走 clawhub search 子命令
func (m *Manager) searchViaCLI(keyword string) ([]models.ClawHubSkill, error) {
	if !m.HasCLI() {
		return nil, fmt.Errorf("clawhub CLI 未安装且网络不可达")
	}
	cmd := exec.Command("clawhub", "search", keyword)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("clawhub search: %w", err)
	}
	return parseSearchOutput(out.String()), nil
}

// loadRegistrySkills 从 GitHub 加载 registry 仓库中所有技能
func (m *Manager) loadRegistrySkills() ([]models.ClawHubSkill, error) {
	m.cache.mu.RLock()
	if m.cache.loaded && time.Since(m.cache.loadedAt) < m.cache.cacheTTL {
		defer m.cache.mu.RUnlock()
		// 返回拷贝
		out := make([]models.ClawHubSkill, len(m.cache.skills))
		copy(out, m.cache.skills)
		return out, nil
	}
	m.cache.mu.RUnlock()

	skills, err := m.fetchAllSkillsFromGitHub()
	if err != nil {
		return nil, err
	}

	m.cache.mu.Lock()
	m.cache.skills = skills
	m.cache.loaded = true
	m.cache.loadedAt = time.Now()
	m.cache.mu.Unlock()

	out := make([]models.ClawHubSkill, len(skills))
	copy(out, skills)
	return out, nil
}

// fetchAllSkillsFromGitHub 通过 GitHub Trees API 列出 registry 中所有 SKILL.md
// 仓库结构：<owner>/<slug>/SKILL.md
func (m *Manager) fetchAllSkillsFromGitHub() ([]models.ClawHubSkill, error) {
	// 1. 拿 default branch
	repoURL := fmt.Sprintf("https://api.github.com/repos/%s", m.registry)
	repoInfo, err := m.httpGetJSON(repoURL)
	if err != nil {
		return nil, fmt.Errorf("获取 registry 信息: %w（请在设置中配置 GitHub Token）", err)
	}
	defaultBranch, _ := repoInfo["default_branch"].(string)
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	// 2. 递归拉取整棵 tree
	treeURL := fmt.Sprintf("https://api.github.com/repos/%s/git/trees/%s?recursive=1", m.registry, defaultBranch)
	treeResp, err := m.httpGetJSON(treeURL)
	if err != nil {
		return nil, fmt.Errorf("获取 tree: %w", err)
	}
	truncated, _ := treeResp["truncated"].(bool)
	if truncated {
		// 仓库过大被截断，仅返回已有数据
	}

	treeArr, _ := treeResp["tree"].([]any)
	// 收集 (owner, slug) 对
	type pair struct{ owner, slug string }
	pairs := map[pair]struct{}{}
	for _, t := range treeArr {
		entry, _ := t.(map[string]any)
		path, _ := entry["path"].(string)
		tp, _ := entry["type"].(string)
		if tp != "blob" {
			continue
		}
		// 匹配形如 <owner>/<slug>/SKILL.md 的路径
		if !strings.HasSuffix(strings.ToLower(path), "/skill.md") {
			continue
		}
		parts := strings.Split(path, "/")
		// 期望 parts: [owner, slug, SKILL.md]，但可能更深（如 owner/slug/sub/SKILL.md）
		// 取最后三段判断
		if len(parts) < 3 {
			continue
		}
		owner := parts[len(parts)-3]
		slug := parts[len(parts)-2]
		pairs[pair{owner, slug}] = struct{}{}
	}

	// 3. 为每个 pair 拉取 SKILL.md 头部信息（name/description/version/tags）
	results := make([]models.ClawHubSkill, 0, len(pairs))
	type fetchResult struct {
		skill models.ClawHubSkill
		err   error
	}
	var wg sync.WaitGroup
	ch := make(chan fetchResult, len(pairs))
	for p := range pairs {
		wg.Add(1)
		go func(p pair) {
			defer wg.Done()
			s, err := m.fetchSkillMetadata(p.owner, p.slug, defaultBranch)
			if err != nil {
				ch <- fetchResult{err: err}
				return
			}
			ch <- fetchResult{skill: s}
		}(p)
	}
	wg.Wait()
	close(ch)
	for r := range ch {
		if r.err == nil {
			results = append(results, r.skill)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Owner != results[j].Owner {
			return results[i].Owner < results[j].Owner
		}
		return results[i].Slug < results[j].Slug
	})
	return results, nil
}

// fetchSkillMetadata 拉取 owner/slug/SKILL.md 并解析头部
func (m *Manager) fetchSkillMetadata(owner, slug, branch string) (models.ClawHubSkill, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s/SKILL.md", m.registry, branch, owner, slug)
	body, err := m.httpGet(url)
	if err != nil {
		return models.ClawHubSkill{}, err
	}
	info := parseSkillFrontMatter(body)
	s := models.ClawHubSkill{
		Owner:       owner,
		Slug:        slug,
		Name:        info.Name,
		Description: info.Description,
		Version:     info.Version,
		Tags:        info.Tags,
		Author:      info.Author,
	}
	if s.Name == "" {
		s.Name = slug
	}
	return s, nil
}

// httpGet GET 请求并返回 body
func (m *Manager) httpGet(url string) (string, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "skills-manager")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	data, _ := io.ReadAll(resp.Body)
	return string(data), nil
}

// httpGetJSON GET 请求并解析为 map
func (m *Manager) httpGetJSON(url string) (map[string]any, error) {
	body, err := m.httpGet(url)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- 元数据解析 ----------

// SkillFrontMatter 是 SKILL.md 头部 YAML 解析结果
type SkillFrontMatter struct {
	Name        string
	Description string
	Version     string
	Tags        []string
	Author      string
}

// parseSkillFrontMatter 解析 SKILL.md 头部的 name:/description:/version:/tags:/author: 字段
func parseSkillFrontMatter(text string) SkillFrontMatter {
	info := SkillFrontMatter{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// 跳过 Markdown 标题
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		// 跳过 markdown 分隔
		if trimmed == "---" {
			continue
		}
		lower := strings.ToLower(trimmed)
		switch {
		case strings.HasPrefix(lower, "name:"):
			info.Name = strings.TrimSpace(trimmed[5:])
		case strings.HasPrefix(lower, "description:"):
			info.Description = strings.TrimSpace(trimmed[12:])
		case strings.HasPrefix(lower, "version:"):
			info.Version = strings.TrimSpace(trimmed[8:])
		case strings.HasPrefix(lower, "author:"):
			info.Author = strings.TrimSpace(trimmed[7:])
		case strings.HasPrefix(lower, "tags:"):
			raw := strings.TrimSpace(trimmed[5:])
			raw = strings.Trim(raw, "[]")
			for _, t := range strings.Split(raw, ",") {
				t = strings.TrimSpace(t)
				t = strings.Trim(t, "\"'")
				if t != "" {
					info.Tags = append(info.Tags, t)
				}
			}
		}
	}
	return info
}

// matchSkill 判断 skill 是否匹配关键词
func matchSkill(s models.ClawHubSkill, kw string) bool {
	fields := []string{
		strings.ToLower(s.Owner),
		strings.ToLower(s.Slug),
		strings.ToLower(s.Name),
		strings.ToLower(s.Description),
		strings.ToLower(s.Author),
	}
	for _, f := range fields {
		if strings.Contains(f, kw) {
			return true
		}
	}
	for _, t := range s.Tags {
		if strings.Contains(strings.ToLower(t), kw) {
			return true
		}
	}
	return false
}

// relevanceScore 返回 skill 与关键词的相关度
func relevanceScore(s models.ClawHubSkill, kw string) int {
	score := 0
	ow := strings.ToLower(s.Owner + "/" + s.Slug)
	if strings.Contains(ow, kw) {
		score += 100
		if ow == kw {
			score += 200
		}
	}
	if strings.Contains(strings.ToLower(s.Name), kw) {
		score += 50
	}
	if strings.Contains(strings.ToLower(s.Description), kw) {
		score += 20
	}
	for _, t := range s.Tags {
		if strings.EqualFold(t, kw) {
			score += 30
		}
	}
	return score
}

// parseSearchOutput 兜底：解析 clawhub CLI 自由文本输出
func parseSearchOutput(text string) []models.ClawHubSkill {
	var out []models.ClawHubSkill
	seen := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		for _, p := range parts {
			if i := strings.Index(p, "/"); i > 0 && i < len(p)-1 {
				owner := p[:i]
				slug := p[i+1:]
				if !isValidOwnerSlug(owner) || !isValidOwnerSlug(slug) {
					continue
				}
				key := owner + "/" + slug
				if seen[key] {
					continue
				}
				seen[key] = true
				out = append(out, models.ClawHubSkill{
					Owner: owner, Slug: slug, Name: key,
				})
			}
		}
	}
	return out
}

func isValidOwnerSlug(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}

// ---------- 安装（拉取技能到 skillspool）----------

// FetchSkill 从 ClawHub 拉取 owner/slug 技能到本地临时目录，返回该目录路径
func (m *Manager) FetchSkill(owner, slug string) (string, error) {
	cacheRoot := filepath.Join(m.skillspoolRoot, ".cache", "clawhub")
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return "", fmt.Errorf("mkdir cache: %w", err)
	}
	tmp, err := os.MkdirTemp(cacheRoot, fmt.Sprintf("%s-%s-", owner, slug))
	if err != nil {
		return "", fmt.Errorf("mktemp: %w", err)
	}

	// 方式 A：CLI
	if m.HasCLI() {
		cmd := exec.Command("clawhub", "install", owner+"/"+slug)
		cmd.Dir = tmp
		var buf strings.Builder
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		if err := cmd.Run(); err == nil {
			if found := findSkillDir(tmp); found != "" {
				return found, nil
			}
		}
	}

	// 方式 B：直接拉 GitHub
	if err := m.fetchSkillFromGitHub(owner, slug, tmp); err != nil {
		return "", fmt.Errorf("拉取技能失败: %w", err)
	}
	if found := findSkillDir(tmp); found != "" {
		return found, nil
	}
	return "", fmt.Errorf("未找到 SKILL.md (owner=%s slug=%s)", owner, slug)
}

// fetchSkillFromGitHub 通过 GitHub Contents API 递归拉取 owner/slug 下所有文件
func (m *Manager) fetchSkillFromGitHub(owner, slug, dst string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s/%s", m.registry, owner, slug)
	return m.fetchContentsRecursive(url, dst)
}

// fetchContentsRecursive 递归拉取 GitHub 目录内容
func (m *Manager) fetchContentsRecursive(apiURL, dstDir string) error {
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "skills-manager")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return fmt.Errorf("registry 中不存在: %s", apiURL)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiURL)
	}
	body, _ := io.ReadAll(resp.Body)
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err != nil {
		return err
	}
	for _, item := range arr {
		typ, _ := item["type"].(string)
		name, _ := item["name"].(string)
		if name == "" || strings.Contains(name, "..") {
			continue
		}
		target := filepath.Join(dstDir, name)
		switch typ {
		case "file":
			dl, _ := item["download_url"].(string)
			if dl == "" {
				continue
			}
			if err := m.downloadFile(dl, target); err != nil {
				return err
			}
		case "dir":
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			subURL, _ := item["url"].(string)
			if subURL == "" {
				continue
			}
			if err := m.fetchContentsRecursive(subURL, target); err != nil {
				return err
			}
		}
	}
	return nil
}

// downloadFile 下载文件
func (m *Manager) downloadFile(url, dst string) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "skills-manager")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	data, _ := io.ReadAll(resp.Body)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// findSkillDir 深度搜索包含 SKILL.md 的目录
func findSkillDir(root string) string {
	var best string
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if _, err2 := os.Stat(filepath.Join(p, "SKILL.md")); err2 == nil {
			best = p
		}
		return nil
	})
	return best
}

// ---------- 安装到 skillspool ----------

// Install 把 ClawHub 技能安装到 skillspool，返回目标目录、解析后的 skillName/version
func (m *Manager) Install(owner, slug string) (targetDir, skillName, version string, err error) {
	tmp, err := m.FetchSkill(owner, slug)
	if err != nil {
		return "", "", "", err
	}
	info, err := lightweightInfo(tmp)
	if err != nil {
		return "", "", "", fmt.Errorf("parse SKILL.md: %w", err)
	}
	skillName = info.Name
	if skillName == "" {
		skillName = slug
	}
	version = info.Version
	if version == "" {
		version = time.Now().Format("2006.01.02")
	}
	targetDir = filepath.Join(m.skillspoolRoot, sanitize(skillName), version)
	if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
		return "", "", "", fmt.Errorf("mkdir target: %w", err)
	}
	if err := copyDirContents(tmp, targetDir); err != nil {
		return "", "", "", fmt.Errorf("copy to skillspool: %w", err)
	}
	return targetDir, skillName, version, nil
}

// ---------- 轻量级 SKILL.md 解析 ----------

// lightweightInfo 仅解析 name/version 字段
func lightweightInfo(dir string) (*models.SkillInfo, error) {
	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return nil, err
	}
	info := parseSkillFrontMatter(string(data))
	return &models.SkillInfo{Name: info.Name, Version: info.Version}, nil
}

// sanitize 把任意字符串变为安全路径片段
func sanitize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "skill"
	}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r)
		case r >= '0' && r <= '9':
			out = append(out, r)
		case r == '-' || r == '_' || r == '.':
			out = append(out, r)
		case r == ' ' || r == '/':
			out = append(out, '-')
		}
	}
	result := string(out)
	if result == "" {
		result = "skill"
	}
	return result
}

// ---------- 目录复制（独立实现避免循环依赖） ----------

func copyDirContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDirR(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			info, _ := e.Info()
			mode := os.FileMode(0o644)
			if info != nil {
				mode = info.Mode()
			}
			if err := os.WriteFile(dstPath, data, mode); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyDirR(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// RawVersion 返回 clawhub --version
func (m *Manager) RawVersion() (string, error) {
	if !m.HasCLI() {
		return "", fmt.Errorf("clawhub CLI 未安装")
	}
	out, err := exec.Command("clawhub", "--version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

var _ = json.Marshal
