package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/internal/clawhub"
	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// MarketSearcher performs aggregated search across multiple skill sources.
type MarketSearcher struct {
	poolPath       string
	marketSources  []models.MarketSource
	clawhub         *clawhub.Manager
	skillsSh        *SkillsShResolver
	githubResolver  *GitHubResolver
	httpResolver    *HTTPResolver
	localResolver   *LocalResolver
}

// NewMarketSearcher creates a new MarketSearcher with its own ClawHub manager.
func NewMarketSearcher(poolPath string, marketSources []models.MarketSource) *MarketSearcher {
	return &MarketSearcher{
		poolPath:       poolPath,
		marketSources:  marketSources,
		clawhub:         clawhub.New(poolPath),
		skillsSh:        NewSkillsShResolver(),
		githubResolver:  &GitHubResolver{},
		httpResolver:    &HTTPResolver{},
		localResolver:   &LocalResolver{},
	}
}

// NewMarketSearcherWithClawHub creates a new MarketSearcher using a pre-existing
// ClawHub manager (for cache reuse across calls).
func NewMarketSearcherWithClawHub(poolPath string, marketSources []models.MarketSource, clawhubMgr *clawhub.Manager) *MarketSearcher {
	return &MarketSearcher{
		poolPath:       poolPath,
		marketSources:  marketSources,
		clawhub:         clawhubMgr,
		skillsSh:        NewSkillsShResolver(),
		githubResolver:  &GitHubResolver{},
		httpResolver:    &HTTPResolver{},
		localResolver:   &LocalResolver{},
	}
}

// SearchAll searches all sources in parallel and returns grouped results.
// Built-in sources (ClawHub, skills.sh, local pool) are always searched.
// Configurable market sources (GitHub repos, registries) are searched if enabled.
func (m *MarketSearcher) SearchAll(ctx context.Context, keyword string) []models.MarketSearchResult {
	if ctx == nil {
		ctx = context.Background()
	}

	type job struct {
		name string
		typ  string
		fn   func() ([]models.MarketSearchSkill, error)
	}

	var jobs []job

	// Built-in: local pool
	jobs = append(jobs, job{
		name: "本地技能池",
		typ:  "pool",
		fn:   func() ([]models.MarketSearchSkill, error) { return m.searchPool(ctx, keyword) },
	})

	// Built-in: ClawHub
	jobs = append(jobs, job{
		name: "ClawHub",
		typ:  "clawhub",
		fn:   func() ([]models.MarketSearchSkill, error) { return m.searchClawHub(ctx, keyword) },
	})

	// Built-in: skills.sh
	jobs = append(jobs, job{
		name: "skills.sh",
		typ:  "skillssh",
		fn:   func() ([]models.MarketSearchSkill, error) { return m.searchSkillsSh(ctx, keyword) },
	})

	// Configurable: market sources
	for _, src := range m.marketSources {
		if !src.Enabled {
			continue
		}
		src := src // capture
		switch src.Type {
		case "pool":
			// Pool is already handled as built-in
			continue
		case "github":
			jobs = append(jobs, job{
				name: src.Name,
				typ:  "github",
				fn:   func() ([]models.MarketSearchSkill, error) { return m.searchGitHubRepo(ctx, src, keyword) },
			})
		case "registry":
			jobs = append(jobs, job{
				name: src.Name,
				typ:  "registry",
				fn:   func() ([]models.MarketSearchSkill, error) { return m.searchRegistry(ctx, src, keyword) },
			})
		}
	}

	// Run all jobs in parallel
	var wg sync.WaitGroup
	results := make([]models.MarketSearchResult, len(jobs))
	for i, j := range jobs {
		wg.Add(1)
		go func(idx int, j job) {
			defer wg.Done()
			// Recover from panics to prevent crashing the entire program.
			defer func() {
				if r := recover(); r != nil {
					results[idx] = models.MarketSearchResult{
						SourceName: j.name,
						SourceType: j.typ,
						Skills:     []models.MarketSearchSkill{},
						Error:      fmt.Sprintf("内部错误: %v", r),
					}
				}
			}()

			// Check if the overall context has already expired.
			if ctx.Err() != nil {
				results[idx] = models.MarketSearchResult{
					SourceName: j.name,
					SourceType: j.typ,
					Skills:     []models.MarketSearchSkill{},
					Error:      "搜索超时，已取消",
				}
				return
			}

			skills, err := j.fn()
			// Ensure skills is never nil to avoid frontend null.map() crash.
			if skills == nil {
				skills = []models.MarketSearchSkill{}
			}
			result := models.MarketSearchResult{
				SourceName: j.name,
				SourceType: j.typ,
				Skills:     skills,
			}
			if err != nil {
				result.Error = err.Error()
			}
			results[idx] = result
		}(i, j)
	}
	wg.Wait()

	// Filter out results with no skills and no error (empty sources)
	var filtered []models.MarketSearchResult
	for _, r := range results {
		if len(r.Skills) > 0 || r.Error != "" {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

// searchPool searches the local skill pool by keyword.
func (m *MarketSearcher) searchPool(ctx context.Context, keyword string) ([]models.MarketSearchSkill, error) {
	entries, err := os.ReadDir(m.poolPath)
	if err != nil {
		return nil, fmt.Errorf("read pool dir: %w", err)
	}

	kw := strings.ToLower(strings.TrimSpace(keyword))
	var skills []models.MarketSearchSkill

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillName := entry.Name()
		// Walk for SKILL.md files
		skillDir := filepath.Join(m.poolPath, skillName)
		skillMD := filepath.Join(skillDir, "SKILL.md")
		data, err := os.ReadFile(skillMD)
		if err != nil {
			// Also check subdirectories (versioned skills)
			subEntries, subErr := os.ReadDir(skillDir)
			if subErr != nil {
				continue
			}
			for _, sub := range subEntries {
				if !sub.IsDir() {
					continue
				}
				subMD := filepath.Join(skillDir, sub.Name(), "SKILL.md")
				subData, subErr := os.ReadFile(subMD)
				if subErr != nil {
					continue
				}
				desc := extractDescription(string(subData))
				if kw == "" || matchSkillField(skillName, desc, kw) {
					skills = append(skills, models.MarketSearchSkill{
						Name:        skillName,
						Namespace:   "pool",
						Version:     sub.Name(),
						Description: desc,
						LocalPath:   filepath.Join(skillDir, sub.Name()),
					})
				}
			}
			continue
		}
		desc := extractDescription(string(data))
		if kw == "" || matchSkillField(skillName, desc, kw) {
			parsed, _ := storage.ParseSkillFile(skillMD)
			version := "latest"
			if parsed.Version != "" {
				version = parsed.Version
			}
			skills = append(skills, models.MarketSearchSkill{
				Name:        skillName,
				Namespace:   "pool",
				Version:     version,
				Description: desc,
				LocalPath:   skillDir,
			})
		}
	}

	return skills, nil
}

// searchClawHub searches ClawHub for skills matching the keyword.
func (m *MarketSearcher) searchClawHub(ctx context.Context, keyword string) ([]models.MarketSearchSkill, error) {
	type result struct {
		skills []models.MarketSearchSkill
		err    error
	}
	ch := make(chan result, 1)
	go func() {
		results, err := m.clawhub.Search(keyword)
		if err != nil {
			ch <- result{err: err}
			return
		}
		skills := make([]models.MarketSearchSkill, 0, len(results))
		for _, r := range results {
			skills = append(skills, models.MarketSearchSkill{
				Name:        r.Name,
				Namespace:   "clawhub:" + r.Owner,
				Version:     r.Version,
				Description: r.Description,
				Source:      r.Owner + "/" + r.Slug,
			})
		}
		ch <- result{skills: skills}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("ClawHub 搜索超时")
	case r := <-ch:
		return r.skills, r.err
	}
}

// searchSkillsSh searches skills.sh for skills matching the keyword.
func (m *MarketSearcher) searchSkillsSh(ctx context.Context, keyword string) ([]models.MarketSearchSkill, error) {
	type result struct {
		skills []models.MarketSearchSkill
		err    error
	}
	ch := make(chan result, 1)
	go func() {
		results, err := m.skillsSh.Search(context.Background(), keyword, 20)
		if err != nil {
			ch <- result{err: err}
			return
		}
		skills := make([]models.MarketSearchSkill, 0, len(results))
		for _, r := range results {
			skills = append(skills, models.MarketSearchSkill{
				Name:        r.Name,
				Namespace:   "skillssh:" + r.Source,
				Version:     fmt.Sprintf("%d", r.Installs),
				Source:      r.ID,
				Installs:    r.Installs,
			})
		}

		// Fetch descriptions from GitHub in parallel
		fetchDescriptionsForSkillsSh(skills, results)

		ch <- result{skills: skills}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("skills.sh 搜索超时")
	case r := <-ch:
		return r.skills, r.err
	}
}

// fetchDescriptionsForSkillsSh fetches SKILL.md descriptions from GitHub for skills.sh results.
func fetchDescriptionsForSkillsSh(skills []models.MarketSearchSkill, rawResults []skillsShSkill) {
	var wg sync.WaitGroup
	for i := range skills {
		if i >= len(rawResults) {
			break
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r := rawResults[idx]
			desc := fetchSkillDescriptionFromGitHub(r.Source, r.Name)
			if desc != "" {
				skills[idx].Description = desc
			}
		}(i)
	}
	wg.Wait()
}

// fetchSkillDescriptionFromGitHub fetches the description field from a skill's SKILL.md on GitHub.
func fetchSkillDescriptionFromGitHub(source, skillName string) string {
	// Sanitize skillName: replace colons with dashes for URL paths
	safeName := strings.ReplaceAll(skillName, ":", "-")

	// Try common path patterns for SKILL.md
	// skills.sh repos typically use: skills/<skillName>/SKILL.md
	// Some repos use: <skillName>/SKILL.md or just SKILL.md at root
	paths := []string{
		fmt.Sprintf("https://raw.githubusercontent.com/%s/main/skills/%s/SKILL.md", source, safeName),
		fmt.Sprintf("https://raw.githubusercontent.com/%s/master/skills/%s/SKILL.md", source, safeName),
		fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s/SKILL.md", source, safeName),
		fmt.Sprintf("https://raw.githubusercontent.com/%s/master/%s/SKILL.md", source, safeName),
		fmt.Sprintf("https://raw.githubusercontent.com/%s/main/SKILL.md", source),
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for _, url := range paths {
		desc, ok := fetchDescriptionFromURL(client, url)
		if ok && desc != "" {
			return desc
		}
	}
	return ""
}

// fetchDescriptionFromURL fetches a SKILL.md from a URL and extracts the description.
func fetchDescriptionFromURL(client *http.Client, url string) (string, bool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", false
	}
	req.Header.Set("User-Agent", "skills-manager")
	resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", false
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}
	return extractDescription(string(data)), true
}

// searchGitHubRepo searches a GitHub repo for skills matching the keyword.
// Uses GitHub Code Search API when keyword is provided, otherwise clones and scans.
func (m *MarketSearcher) searchGitHubRepo(ctx context.Context, src models.MarketSource, keyword string) ([]models.MarketSearchSkill, error) {
	sourceURL := src.URL
	branch := src.Branch
	if branch == "" {
		branch = "main"
	}

	kw := strings.TrimSpace(keyword)
	if kw != "" {
		// Use GitHub Code Search API for keyword search
		return m.searchGitHubViaAPI(ctx, sourceURL, kw)
	}

	// No keyword: clone and scan (original behavior)
	opts := ResolveOptions{Ref: branch}
	resolved, err := m.githubResolver.Resolve(ctx, sourceURL, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		for _, r := range resolved {
			if r.Cleanup != nil {
				r.Cleanup()
			}
		}
	}()

	skills := make([]models.MarketSearchSkill, 0, len(resolved))
	for _, r := range resolved {
		skills = append(skills, models.MarketSearchSkill{
			Name:      r.Name,
			Namespace: r.Namespace,
			Version:   r.Version,
		})
	}
	return skills, nil
}

// searchGitHubViaAPI uses the GitHub Code Search API to find SKILL.md files matching the keyword.
func (m *MarketSearcher) searchGitHubViaAPI(ctx context.Context, sourceURL, keyword string) ([]models.MarketSearchSkill, error) {
	ownerRepo := parseGitHubOwnerRepo(sourceURL)
	if ownerRepo == "" {
		return nil, fmt.Errorf("invalid GitHub source: %s", sourceURL)
	}

	// GitHub Code Search API: search for SKILL.md files containing the keyword
	// Format: q=<keyword>+in:file+path:SKILL.md+repo:<owner/repo>
	searchURL := fmt.Sprintf(
		"https://api.github.com/search/code?q=%s+in:file+path:SKILL.md+repo:%s",
		keyword, ownerRepo,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create GitHub search request: %w", err)
	}
	req.Header.Set("User-Agent", "skills-manager")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub search API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub search API returned status %d", resp.StatusCode)
	}

	var searchResp struct {
		Items []struct {
			Path       string `json:"path"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("parse GitHub search response: %w", err)
	}

	// Extract skill names from paths: owner/slug/SKILL.md -> slug
	skills := make([]models.MarketSearchSkill, 0, len(searchResp.Items))
	seen := map[string]bool{}
	for _, item := range searchResp.Items {
		parts := strings.Split(item.Path, "/")
		if len(parts) < 2 {
			continue
		}
		slug := parts[len(parts)-2]
		if seen[slug] {
			continue
		}
		seen[slug] = true
		skills = append(skills, models.MarketSearchSkill{
			Name:      slug,
			Namespace: "github:" + item.Repository.FullName,
			Version:   "latest",
			Source:    item.Repository.FullName,
		})
	}
	return skills, nil
}

// searchRegistry searches an HTTP registry for skills matching the keyword.
func (m *MarketSearcher) searchRegistry(ctx context.Context, src models.MarketSource, keyword string) ([]models.MarketSearchSkill, error) {
	if !m.httpResolver.CanHandle(src.URL) {
		return nil, fmt.Errorf("unsupported registry URL: %s", src.URL)
	}

	opts := ResolveOptions{}
	resolved, err := m.httpResolver.Resolve(ctx, src.URL, opts)
	if err != nil {
		return nil, err
	}

	kw := strings.ToLower(strings.TrimSpace(keyword))
	skills := make([]models.MarketSearchSkill, 0, len(resolved))
	for _, r := range resolved {
		if kw == "" || strings.Contains(strings.ToLower(r.Name), kw) || strings.Contains(strings.ToLower(r.Namespace), kw) {
			skills = append(skills, models.MarketSearchSkill{
				Name:      r.Name,
				Namespace: r.Namespace,
				Version:   r.Version,
			})
		}
	}
	return skills, nil
}

// matchKeyword checks if the keyword appears in the skill content.
// Deprecated: use matchSkillField for more precise matching.
func matchKeyword(content, keyword string) bool {
	lower := strings.ToLower(content)
	return strings.Contains(lower, keyword)
}

// matchSkillField checks if the keyword matches the skill name or description.
// This avoids matching common words inside the full SKILL.md body.
func matchSkillField(name, description, keyword string) bool {
	kw := strings.ToLower(keyword)
	if strings.Contains(strings.ToLower(name), kw) {
		return true
	}
	if strings.Contains(strings.ToLower(description), kw) {
		return true
	}
	return false
}

// extractDescription extracts a short description from SKILL.md content.
func extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "description:") {
			return strings.TrimSpace(line[12:])
		}
	}
	// Fallback: first non-empty line after the title
	started := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "---" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			started = true
			continue
		}
		if started && line != "" {
			if len(line) > 120 {
				line = line[:120] + "..."
			}
			return line
		}
	}
	return ""
}