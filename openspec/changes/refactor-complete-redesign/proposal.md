# Change: Skills Manager Complete Redesign

## Why

The current codebase was built incrementally without following the `reDesign.md` architecture. Key problems:
- `backend/internal/` modules are tightly coupled with unclear responsibilities (lifecycle/agent/installer overlap)
- No separation between "skill source discovery", "skill storage", and "skill distribution" domains
- Storage directory (`skillspool/`) doesn't support namespacing or versioning
- Frontend only has 3 pages, missing Market and Settings
- CLI commands are incomplete (no search, config, doctor, export/import)
- Lock file and versioning system missing

A complete rewrite is needed to align with the DDD architecture defined in `reDesign.md`.

## What Changes

- **Brand new Go backend** organized into 4 DDD domains: source/storage/distribute/operations
- **New data models**: Config, Index, LockFile with proper JSON schemas
- **New storage layout**: `~/.skill-repo/skills/{namespace}/{name}@{version}/`
- **New CLI binary** (`skill`) with 16 subcommands using Cobra
- **New frontend** with 5 pages (Dashboard/Market/Skills/Detail/Settings)
- **Resolver pattern** for multi-source skill discovery (GitHub/HTTP/local/ZIP)
- **Installer orchestration** with symlink/copy/fallback modes
- **Lock file tracking** for installed skills per agent
- All of this is a **complete rewrite** — the old `backend/` code will be replaced

## Impact

- Affected specs: skill-source, skill-storage, skill-distribute, skill-operations, cli-commands, frontend-gui
- Affected code: all backend files under `backend/`, all frontend files under `frontend/`, Wails entry points
- Old `backend/internal/` code (lifecycle/agent/installer/config/skill) will be removed
- Old `backend/cmd/sm/` will be replaced by `backend/cmd/skill/`