# Proposal: Skill Cleanup, Version Management & Enhanced Features

## Why: Problem Statement

As users install more skills across multiple agents, several pain points emerge:

1. **Project Skill Bloat**: Project directories accumulate skills over time, making `.agent/skills` directories cluttered and unmanageable
2. **Orphaned Symlinks**: When skills are deleted from the global library, project symlinks become broken but remain undetected
3. **Version Chaos**: Multiple versions of the same skill accumulate without easy management
4. **No Health Visibility**: Users can't tell if their skill installation is healthy or corrupted
5. **No Usage Insights**: No visibility into which skills are actually being used by which agents

## What: Feature Summary

This change adds comprehensive skill lifecycle management:

| Feature | Description |
|---------|-------------|
| **Skill Cleanup** | Uninstall skills from projects, clean orphaned symlinks, batch清理 |
| **Version Management** | List/switch/delete skill versions |
| **Health Check** | Detect broken symlinks, missing files, version conflicts |
| **Usage Statistics** | Track which agents have installed which skills |

## Impact

- **Positive**: Cleaner projects, healthier skill ecosystem, better observability
- **Risk**: Low - all operations are additive, no destructive defaults
- **Migration**: No breaking changes to existing functionality

## Out of Scope

- Skill dependencies analysis (complex, low priority)
- Import/export (can be added later)
- Batch update (requires market API integration)
