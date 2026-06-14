---
name: "dev2del_workflow"
description: "Development-to-Delivery quality gate workflow requiring test-first design (≥99% coverage), 100% test pass, and code review loops. Invoke when starting a multi-step feature, before claiming work complete, or when user asks for quality assurance process."
---

# Dev2Del Workflow

A rigorous quality gate workflow for development-to-delivery, ensuring each phase is validated before proceeding.

## Workflow Diagram

```
测试用例设计 → Review覆盖率(≥99%)
       │
    [通过]
       │
开发实现 → 测试(100%通过) → Code Review
    ↑                            │
    └────── 有改动需修改 ──────┘
               │
          [无改动 = 完成]
```

## Phase 1: Test Case Design

**Before writing any implementation code**, design test cases first:

1. Identify all key scenarios: happy path, error paths, edge cases, boundary conditions
2. For each scenario, describe:
   - Input conditions
   - Expected behavior/assertions
   - Why this scenario matters
3. Review test coverage — must achieve **≥99%**
4. For any scenario that cannot be covered, document the reason explicitly

### Coverage Targets

| Area | Target | Notes |
|------|--------|-------|
| Backend (Go) `go test ./... -cover` | ≥99% | Use `t.TempDir()` for filesystem tests, table-driven tests for variations |
| Frontend (TypeScript) | Critical paths | Focus on type safety via `tsc --noEmit`, logic in hooks/components |
| Integration | Key flows | Agent detection, config save/load, pool scanning, skill install |

## Phase 2: Development Implementation

Implement the code following these principles:

1. **Test-driven**: Let test cases guide implementation
2. **Avoid over-engineering**: Only implement what's needed for passing tests
3. **Prefer edits over new files**: Modify existing files where possible
4. **Config over code**: Make behavior configurable rather than hardcoded
5. **Snake_case JSON**: TypeScript interfaces must use snake_case to match Go's `json` tags
6. **Never return nil slices**: Always return initialized empty slices (`make([]T, 0)`) instead of nil

## Phase 3: Testing (100% Pass)

Run all tests and verify:

```bash
# Backend
go test ./... -count=1          # All packages
go test ./... -cover             # Coverage report
go build ./...                   # Build check

# Frontend
cd frontend && npx tsc --noEmit  # TypeScript check
cd frontend && npm run build     # Build check

# Full build
wails build                      # Desktop app build
```

- All tests must pass (exit code 0)
- Coverage should be ≥99%
- TypeScript must compile with zero errors

## Phase 4: Code Review

Use the `requesting-code-review` skill to evaluate the work:

1. Get git context: `BASE_SHA=$(git rev-parse HEAD~1)` and `HEAD_SHA=$(git rev-parse HEAD)`
2. Dispatch code reviewer via Task tool with the `code-reviewer.md` template
3. Review feedback is categorized as:
   - **Critical** (must fix immediately)
   - **Important** (fix before proceeding)
   - **Minor** (note for later)

## Phase 5: Loop or Deliver

Based on code review results:

- **No changes needed** → Phase complete, proceed to next milestone
- **Changes needed** (Critical or Important) → Return to Phase 2 (Development), make fixes, then re-run Phase 3 (Testing) → Phase 4 (Code Review) again
- Continue looping until Phase 4 returns "no changes needed"

## Engineering Principles

1. **Core first**: "First ensure usability" — implement core functionality before UI polish
2. **Test-first**: Design tests before writing implementation code
3. **Quality gate**: Each phase has an explicit gate that must pass
4. **Loop don't skip**: If code review finds issues, don't skip to delivery — loop back and fix properly
5. **Evidence before assertions**: Run verification commands and confirm output before claiming success