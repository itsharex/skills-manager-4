# Skills Manager — OpenSpec Project

## Project Overview
Skills Manager is an AI Agent skill installation tool. It discovers skills from multiple sources (GitHub/HTTP Registry/local), stores them in a centralized versioned repository, and distributes them to AI Agents via symlink or copy.

## Tech Stack
- Desktop framework: **Wails v2.12.0**
- Backend: **Go 1.22+**
- CLI: **Cobra**
- Frontend: **React 18 + TypeScript + Vite 5**
- UI: **shadcn/ui + Tailwind CSS**
- State: React Context + useReducer

## Directory Conventions
- Go package path: `github.com/username/skillsmanager`
- Backend code: `backend/`
- Frontend code: `frontend/src/`
- Wails entry: project root (`main.go`, `app.go`, `embed.go`)
- CLI binary name: `skill`

## Module Structure (DDD)
- `backend/internal/source/` — Skill discovery (github/http/local resolvers + validator)
- `backend/internal/storage/` — Repository, index, lock, parser, version
- `backend/internal/distribute/` — Installer, symlink, copy, sync, agent
- `backend/internal/operations/` — Health, cleanup, stats, config
- `backend/pkg/models/` — Shared data types
- `backend/pkg/waillib/` — Wails bridge layer (desktop app API)

## Capabilities
| Capability | Package | Description |
|---|---|---|
| `skill-source` | `internal/source/` | Multi-source skill discovery and validation |
| `skill-storage` | `internal/storage/` | Local repository management with index/lock/version |
| `skill-distribute` | `internal/distribute/` | Install, uninstall, sync skills to agents |
| `skill-operations` | `internal/operations/` | Health check, cleanup, stats, config |
| `cli-commands` | `backend/cmd/skill/` | CLI interface with 16 subcommands |
| `frontend-gui` | `frontend/` | Wails GUI with 5 pages |

## Naming Conventions
- Change IDs: `refactor-*`, `add-*`, `update-*`, `fix-*` (kebab-case, verb-led)
- Go files: snake_case.go
- Go types: PascalCase
- Go functions: camelCase (unexported), PascalCase (exported)
- Frontend files: PascalCase.tsx
- Frontend functions: camelCase

## Storage Directory
- Repository root: `~/.skill-repo/`
- Store path: `~/.skill-repo/skills/{namespace}/{name}@{version}/`
- Index: `~/.skill-repo/index.json`
- Lock: `~/.skill-repo/lock.json`
- Config: `~/.skill-repo/config.json`

## Installation Modes
- Default: Symlink (`ln -sfn`)
- Fallback: Copy (Windows or permission-denied)
- Forced: `--copy` flag