package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

func (r *HTTPResolver) CanHandle(source string) bool {
	// Handle http:// or https:// that is NOT a GitHub URL
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Exclude GitHub URLs — those are handled by GitHubResolver
		if strings.Contains(source, "github.com") {
			return false
		}
		return true
	}
	// Handle registry:name format
	if strings.HasPrefix(source, "registry:") {
		return true
	}
	return false
}

func (r *HTTPResolver) Resolve(ctx context.Context, source string, opts ResolveOptions) ([]models.ResolvedSkill, error) {
	// Handle registry:name format
	if strings.HasPrefix(source, "registry:") {
		return nil, fmt.Errorf("registry name resolution requires configuration; use a full URL instead: %s", source)
	}

	// Validate that source is an HTTP/HTTPS URL
	if !strings.HasPrefix(source, "http://") && !strings.HasPrefix(source, "https://") {
		return nil, fmt.Errorf("invalid HTTP source: %s", source)
	}

	// Fetch registry index
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch registry index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Parse registry index (JSON array of skill entries)
	var entries []registryEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("parse registry index: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("registry index is empty")
	}

	// Build namespace from source URL
	namespace := "registry:" + source

	skills := make([]models.ResolvedSkill, 0, len(entries))
	for _, entry := range entries {
		version := entry.Version
		if version == "" {
			version = "latest"
		}
		skills = append(skills, models.ResolvedSkill{
			LocalPath: "", // HTTP resolver doesn't download; installer uses URLs
			Namespace: namespace,
			Name:      entry.Name,
			Version:   version,
			Cleanup:   nil, // no cleanup needed
		})
	}

	return skills, nil
}

// registryEntry represents a single entry in a registry index.
type registryEntry struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	URL         string   `json:"url"`
	Tags        []string `json:"tags"`
}