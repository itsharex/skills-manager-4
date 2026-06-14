package distribute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// SyncResult records the result of syncing a skill to an agent.
type SyncResult struct {
	SkillName string `json:"skill_name"`
	AgentID   string `json:"agent_id"`
	Mode      string `json:"mode"` // "symlink", "copy", "skipped"
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// SyncSummary holds the overall result of a sync operation.
type SyncSummary struct {
	Total   int           `json:"total"`
	Success int           `json:"success"`
	Failed  int           `json:"failed"`
	Results []SyncResult  `json:"results"`
}

// SyncSkillToAgent installs a skill (from the repository) to a single agent.
// It creates either a symlink or a copy depending on the mode and platform.
func SyncSkillToAgent(skillPath, agentID string, forceCopy bool) (*SyncResult, error) {
	agentSkillsDir, err := GetAgentSkillsDir(agentID)
	if err != nil {
		return &SyncResult{AgentID: agentID, Success: false, Error: err.Error()}, err
	}

	// Ensure agent skills directory exists
	if err := os.MkdirAll(agentSkillsDir, 0o755); err != nil {
		return &SyncResult{AgentID: agentID, Success: false, Error: err.Error()}, err
	}

	// Determine the link/copy name from the skill path
	skillName := filepath.Base(skillPath)
	linkPath := filepath.Join(agentSkillsDir, skillName)

	// Try symlink first
	mode, err := CreateSymlink(skillPath, linkPath, forceCopy)
	if err != nil {
		// If symlink failed and we're configured to fall back, try copy
		if strings.Contains(err.Error(), "symlink") && !forceCopy {
			if err := CopySkill(skillPath, linkPath); err != nil {
				return &SyncResult{
					AgentID: agentID, SkillName: skillName,
					Success: false, Error: fmt.Sprintf("symlink+copy fallback failed: %v", err),
				}, err
			}
			mode = "copy"
		} else {
			return &SyncResult{
				AgentID: agentID, SkillName: skillName,
				Success: false, Error: err.Error(),
			}, err
		}
	}

	// If symlink returned "copy" mode (forceCopy=true or fallback needed)
	if mode == "copy" {
		if err := CopySkill(skillPath, linkPath); err != nil {
			return &SyncResult{
				AgentID: agentID, SkillName: skillName,
				Success: false, Error: fmt.Sprintf("copy failed: %v", err),
			}, err
		}
	}

	return &SyncResult{
		SkillName: skillName,
		AgentID:   agentID,
		Mode:      mode,
		Success:   true,
	}, nil
}

// UnsyncSkillFromAgent removes a skill installation from an agent.
func UnsyncSkillFromAgent(skillName, agentID string) error {
	agentSkillsDir, err := GetAgentSkillsDir(agentID)
	if err != nil {
		return err
	}

	linkPath := filepath.Join(agentSkillsDir, skillName)

	if IsSymlink(linkPath) {
		return RemoveSymlink(linkPath)
	}

	// Check if it's a copied directory
	info, err := os.Stat(linkPath)
	if os.IsNotExist(err) {
		return nil // already removed
	}
	if err != nil {
		return fmt.Errorf("access agent skill path: %w", err)
	}

	if info.IsDir() {
		return os.RemoveAll(linkPath)
	}

	return os.Remove(linkPath)
}

// SyncSkillsToAgents syncs multiple skills to multiple agents.
// Each skill is identified by its full repository path.
func SyncSkillsToAgents(skillPaths []string, agentIDs []string, forceCopy bool) (*SyncSummary, error) {
	summary := &SyncSummary{
		Total:   len(skillPaths) * len(agentIDs),
		Results: make([]SyncResult, 0, len(skillPaths)*len(agentIDs)),
	}

	for _, skillPath := range skillPaths {
		skillName := filepath.Base(skillPath)
		for _, agentID := range agentIDs {
			result, err := SyncSkillToAgent(skillPath, agentID, forceCopy)
			if err != nil {
				summary.Failed++
			} else {
				summary.Success++
			}
			result.SkillName = skillName
			summary.Results = append(summary.Results, *result)
		}
	}

	return summary, nil
}

// SyncIndexEntry syncs a single index entry to multiple agents.
func SyncIndexEntry(entry models.IndexEntry, repo *models.RepoPaths, agentIDs []string, forceCopy bool) (*SyncSummary, error) {
	skillPath := filepath.Join(repo.SkillsDir, entry.Namespace, fmt.Sprintf("%s@%s", entry.Name, entry.Latest))
	return SyncSkillsToAgents([]string{skillPath}, agentIDs, forceCopy)
}

// SyncAllInstalled syncs all skills recorded in the lock file to their respective agents.
func SyncAllInstalled(lock *models.LockFile, repo *models.RepoPaths, forceCopy bool) (*SyncSummary, error) {
	if lock == nil {
		return &SyncSummary{}, nil
	}

	var allResults []SyncResult
	totalCount := 0
	successCount := 0
	failedCount := 0

	for _, entry := range lock.Skills {
		skillPath := filepath.Join(repo.SkillsDir, entry.SkillID.Namespace, fmt.Sprintf("%s@%s", entry.SkillID.Name, entry.SkillID.Version))
		agentIDs := make([]string, len(entry.Agents))
		for i, agent := range entry.Agents {
			agentIDs[i] = agent.AgentID
		}

		summary, _ := SyncSkillsToAgents([]string{skillPath}, agentIDs, forceCopy)
		successCount += summary.Success
		failedCount += summary.Failed
		totalCount += len(agentIDs)
		allResults = append(allResults, summary.Results...)
	}

	return &SyncSummary{
		Total:   totalCount,
		Success: successCount,
		Failed:  failedCount,
		Results: allResults,
	}, nil
}