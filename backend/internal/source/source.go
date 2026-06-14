package source

import (
	"context"
	"fmt"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// ResolveOptions configures how a source is resolved.
type ResolveOptions struct {
	SubPath string // subdirectory within a repository
	Version string // specific version to resolve
	Ref     string // git branch or tag
}

// Resolver defines the interface for resolving skills from a source.
type Resolver interface {
	// Resolve fetches skills from the given source and returns them.
	Resolve(ctx context.Context, source string, opts ResolveOptions) ([]models.ResolvedSkill, error)
	// CanHandle returns true if this resolver can handle the given source.
	CanHandle(source string) bool
}

// NewResolver returns the appropriate resolver for the given source string.
// It iterates through registered resolvers and returns the first match.
func NewResolver(source string) (Resolver, error) {
	resolvers := []Resolver{
		&GitHubResolver{}, // will be implemented in github.go
		&HTTPResolver{},   // will be implemented in http.go
		&LocalResolver{},  // will be implemented in local.go
	}
	for _, r := range resolvers {
		if r.CanHandle(source) {
			return r, nil
		}
	}
	return nil, fmt.Errorf("no resolver found for source: %s", source)
}

// --- Resolver type definitions (implemented in separate files) ---

// GitHubResolver handles GitHub repository sources.
type GitHubResolver struct{}

// HTTPResolver handles HTTP registry sources.
type HTTPResolver struct{}

// LocalResolver handles local filesystem sources.
type LocalResolver struct{}