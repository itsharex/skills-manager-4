package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SkillsShResolver resolves skills from the skills.sh registry.
type SkillsShResolver struct {
	client *http.Client
}

// NewSkillsShResolver creates a new SkillsShResolver.
func NewSkillsShResolver() *SkillsShResolver {
	return &SkillsShResolver{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

const skillsShSearchURL = "https://skills.sh/api/search"

// skillsShSkill represents a skill returned by the skills.sh API.
type skillsShSkill struct {
	ID       string `json:"id"`       // owner/repo/slug
	Name     string `json:"name"`
	Installs int    `json:"installs"`
	Source   string `json:"source"` // owner/repo
}

// skillsShResponse is the API response from skills.sh
type skillsShResponse struct {
	Skills []skillsShSkill `json:"skills"`
}

// Search searches skills.sh for skills matching the given keyword.
func (r *SkillsShResolver) Search(ctx context.Context, keyword string, limit int) ([]skillsShSkill, error) {
	if limit <= 0 {
		limit = 20
	}
	queryURL := fmt.Sprintf("%s?q=%s&limit=%d", skillsShSearchURL, url.QueryEscape(keyword), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "skills-manager")
	req.Header.Set("Accept", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("skills.sh search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("skills.sh returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result skillsShResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parse skills.sh response: %w", err)
	}

	return result.Skills, nil
}