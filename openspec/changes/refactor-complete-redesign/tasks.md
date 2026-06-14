## 1. Phase 1: Foundation (1.5 days)
- [x] 1.1 Initialize new project directory structure
- [x] 1.2 Configure go.mod, wails.json, package.json
- [x] 1.3 Implement `pkg/models/models.go` (SkillID, Config, Index, LockFile, LinkTarget, RepoSource)
- [x] 1.4 Implement `internal/operations/config.go` (config read/write, defaults)
- [x] 1.5 Implement `internal/storage/repository.go` (repo init, path generation, init flow)

## 2. Phase 2: Skill Storage Domain (3 days)
- [x] 2.1 Complete `repository.go`: Store/Remove/UpdateLatest methods
- [x] 2.2 Implement `internal/storage/index.go`: index.json CRUD (Add/Remove/Get/List/Update)
- [x] 2.3 Implement `internal/storage/lock.go`: lock.json CRUD (Track/Untrack/GetBySkill/GetByAgent)
- [x] 2.4 Implement `internal/storage/parser.go`: SKILL.md frontmatter + markdown parsing
- [x] 2.5 Implement `internal/storage/version.go`: version comparison and sorting

## 3. Phase 3: Skill Source Domain (3 days)
- [x] 3.1 Implement `internal/source/source.go`: Resolver interface, ResolvedSkill, factory NewResolver
- [x] 3.2 Implement `internal/source/github.go`: GitHub repo detection, shallow clone, skill discovery
- [x] 3.3 Implement `internal/source/http.go`: HTTP registry fetch, skills.sh support
- [x] 3.4 Implement `internal/source/local.go`: local path, folder, ZIP import
- [x] 3.5 Implement `internal/source/validator.go`: SKILL.md format validation

## 4. Phase 4: Skill Distribution Domain (3 days)
- [x] 4.1 Implement `internal/distribute/installer.go`: Install/Uninstall orchestration
- [x] 4.2 Implement `internal/distribute/symlink.go`: symlink creation + Windows fallback
- [x] 4.3 Implement `internal/distribute/copy.go`: copy-mode installation
- [x] 4.4 Implement `internal/distribute/sync.go`: multi-agent bulk sync
- [x] 4.5 Implement `internal/distribute/agent.go`: agent auto-detection and configuration

## 5. Phase 5: Operations Domain (2 days)
- [x] 5.1 Implement `internal/operations/health.go`: doctor diagnostic checks
- [x] 5.2 Implement `internal/operations/cleanup.go`: orphaned symlink/skill cleanup
- [x] 5.3 Implement `internal/operations/stats.go`: statistics collection

## 6. Phase 6: CLI Commands (3 days)
- [x] 6.1 Set up Cobra root command in `backend/cmd/skill/main.go`
- [x] 6.2 Implement `cmd_init.go`: skill init
- [x] 6.3 Implement `cmd_config.go`: skill config get/set + skill repo add/remove/list
- [x] 6.4 Implement `cmd_search.go`: skill search multi-source
- [x] 6.5 Implement `cmd_list.go`: skill list with filtering
- [x] 6.6 Implement `cmd_info.go`: skill info + skill show + skill validate
- [x] 6.7 Implement `cmd_install.go`: skill install with interactive multi-select
- [x] 6.8 Implement `cmd_uninstall.go`: skill uninstall
- [x] 6.9 Implement `cmd_update.go`: skill update
- [x] 6.10 Implement `cmd_sync.go`: skill sync
- [x] 6.11 Implement `cmd_edit.go`: skill edit
- [x] 6.12 Implement `cmd_doctor.go`: skill doctor
- [x] 6.13 Implement `cmd_export.go`: skill export + skill import
- [x] 6.14 Implement `cmd_stats.go`: skill stats

## 7. Phase 7: Frontend GUI (4 days)
- [x] 7.1 Scaffold Vite + React + TypeScript + Tailwind + shadcn/ui
- [x] 7.2 Set up routing and sidebar layout (App.tsx)
- [x] 7.3 Implement Dashboard page (metric cards, skill/agent overview)
- [x] 7.4 Implement Market page (search, install from source)
- [x] 7.5 Implement Skills page (installed skills list, detail navigation)
- [x] 7.6 Implement SkillDetail page (metadata display, version tabs)
- [x] 7.7 Implement Settings page (Agent list, config display)
- [x] 7.8 Implement shared components (Button, Card, Badge, Tabs)
- [x] 7.9 Implement hooks/bridge (bridge.ts with Wails + mock fallbacks)
- [x] 7.10 Set up Wails binding bridge (bridge.ts, waillib package)
- [x] 7.11 Set up `pkg/waillib/app.go` unified API facade

## 8. Phase 8: Testing & Release (3 days)
- [x] 8.1 Write unit tests for `internal/source/` (≥80% coverage → 84.8%)
- [x] 8.2 Write unit tests for `internal/storage/` (≥80% coverage → 90.0%)
- [x] 8.3 Write unit tests for `internal/distribute/` (≥80% coverage → 80.3%)
- [x] 8.4 Write unit tests for `internal/operations/` (≥80% coverage → 85.3%)
- [x] 8.5 Write integration tests (GitHub mock, HTTP mock, full install flow)
- [x] 8.6 Write CLI E2E tests (smoke test all 16 commands)
- [x] 8.7 Build + package with Wails (darwin/arm64 binary)