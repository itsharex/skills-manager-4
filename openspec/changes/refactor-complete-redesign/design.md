## Context

Complete rewrite of Skills Manager following `reDesign.md`. The full design document is at `docs/superpowers/specs/2026-06-11-skills-manager-redesign.md`.

## Goals / Non-Goals

- **Goals:**
  - Clean DDD module separation (source/storage/distribute/operations)
  - Versioned, namespaced skill storage at `~/.skill-repo/`
  - Lock-file-tracked installations per AI Agent
  - 16 CLI subcommands covering full lifecycle
  - 5-page Wails GUI with Wails binding
  - GitHub/HTTP registry/local ZIP multi-source support

- **Non-Goals:**
  - Not changing the tech stack (staying with Wails + Go + React + TypeScript)
  - Not implementing MCP/HTTP API server (reserved for future)
  - Not implementing plugin system for custom resolvers

## Decisions

- **Decision: Four-domain DDD layout** (`source`/`storage`/`distribute`/`operations`)
  - Rationale: Each domain has minimal cross-dependency; source discovers, storage persists, distribute places, operations maintains
- **Decision: Resolver interface pattern for skill sources**
  - Rationale: New sources can be added without modifying existing code
- **Decision: `~/.skill-repo/` as storage root with `{namespace}/{name}@{version}` path scheme**
  - Rationale: Supports multi-source coexistence (github:org/repo vs registry:name) and multi-version installs
- **Decision: CLI + Wails share `pkg/api.API` facade**
  - Rationale: Single entry point for all business logic; CLI and GUI are symmetric consumers
- **Decision: Soft deletion for index entries (keep version history)**
  - Rationale: Enables rollback and `skill doctor` repair

## Risks / Trade-offs

- **Risk**: ZIP extraction permissions may vary across OS → Mitigation: use `mholt/archiver` with error fallback
- **Risk**: Git clone for large repos can timeout → Mitigation: `--depth 1` shallow clone with configurable timeout

## Implementation Plan

8 phases totaling ~22.5 days (see tasks.md for detailed checklist)