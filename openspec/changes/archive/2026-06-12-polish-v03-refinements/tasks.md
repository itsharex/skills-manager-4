## 质量门禁规则

每个阶段必须遵循以下循环，达成后方可进入下一阶段：

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

- **测试用例设计**：在开发实现前，先明确测试策略，列出关键测试场景
- **Review 覆盖率**：测试用例覆盖率必须 ≥ 99%，未能覆盖的场景需记录原因（如"UI 交互类无法单元测试"）
- **测试必须 100% 通过**：后端 `go test ./... -cover` 全部通过且覆盖率达标，前端 `npm run build` 通过
- **Code Review** 由 requesting-code-review 技能执行
- Review 后若无改动 → 该阶段完成
- Review 后若有改动 → 回到开发实现，再次循环

---

## 1. 移除 mock 数据

### 测试用例设计
- [ ] 1.1 测试场景：
  - 无 Wails 环境时调用 API → 抛错而非返回假数据
  - 有 Wails 环境时调用 API → 正常调用后端
  - `hasWails()` 在不同浏览器环境下的行为
- [ ] 1.2 Review 覆盖率：≥ 99%，无法覆盖的用例说明原因

### 开发实现
- [ ] 1.3 删除 `bridge.ts` 中的 `mockConfig`, `mockSkills`, `mockAgents`, `mockSearch`, `mockInstall` 函数
- [ ] 1.4 将所有 API 函数的 fallback 改为直接抛错 `throw new Error("Wails backend not available")`

### 测试验证
- [ ] 1.5 `cd frontend && npm run build` 通过，无 mock 数据残留
- [ ] 1.6 `go build ./...` 通过

### Code Review & 循环
- [ ] 1.7 Code Review → 无改动则完成 → 有改动则回到 1.3

---

## 2. 扩展 agent 列表

### 测试用例设计
- [ ] 2.1 测试场景：
  - `KnownAgents()` 返回总数 ≥ 40
  - 各 agent 的 SkillsDir 路径格式正确（均为绝对路径）
  - `DetectAgents()` 返回的 detected agent 数量 ≤ `KnownAgents()` 数量
  - 无重复 ID 或 Name
  - 边界：home 目录不可读时的降级行为
- [ ] 2.2 Review 覆盖率：≥ 99%，无法覆盖的用例说明原因

### 开发实现
- [ ] 2.3 在 `backend/internal/distribute/agent.go` `KnownAgents()` 中补充 agent
  - `bolt` → Bolt (Supermaven)
  - `zed` → Zed AI
  - `zed-dev` → Zed Developer
  - `jupyter` → Jupyter AI
  - `goose` → Google Goose
  - `pony` → Pony AI
  - `replit` → Replit AI
  - `replicant` → Replicant
  - `marsha` → Marsha
  - `qodo` → Qodo (CodiumAI)
  - `cosine` → Cosine
  - `grok` → Grok (xAI)
  - `diffra` → Diffra
  - `codecomet` → CodeComet
  - `softgen` → SoftGen
- [ ] 2.4 确保总数 ≥ 40 个

### 测试验证
- [ ] 2.5 `go test ./backend/internal/distribute/... -cover` 通过，覆盖率 ≥ 99%
- [ ] 2.6 `go build ./...` 通过
- [ ] 2.7 确认 `KnownAgents()` 返回 ≥ 40 条，无重复 ID

### Code Review & 循环
- [ ] 2.8 Code Review → 无改动则完成 → 有改动则回到 2.3

---

## 3. Agent 分组展示增强

### 测试用例设计
- [ ] 3.1 测试场景：
  - 同路径 agent 合并为一行，displayName 为最短名
  - 全部未检测时 displayName = "最短名 等"
  - tooltipName 包含全部名字逗号拼接
  - 已检测组排前面，未检测组排后面
  - 单个 agent 独立路径时不截断
  - 边界：agents 数组为空的 Edge case
  - 边界：所有 agent 检测状态一致时的排序稳定性
- [ ] 3.2 Review 覆盖率：≥ 99%，无法覆盖的用例说明原因

### 开发实现
- [ ] 3.3 `frontend/src/types.ts` 中增强 `AgentGroup` 类型，添加 `displayName: string`, `tooltipName: string`
- [ ] 3.4 `SettingsPage.tsx` 中 `groupAgentsByPath` 返回 `AgentGroup[]`：
  - 同路径 agent 合并
  - displayName = 最短名（无检测时加" 等"）
  - tooltipName = 全部名逗号拼接
- [ ] 3.5 已检测组排前面，未检测组排后面

### 测试验证
- [ ] 3.6 `cd frontend && npx tsc --noEmit` 通过
- [ ] 3.7 `cd frontend && npm run build` 通过

### Code Review & 循环
- [ ] 3.8 Code Review → 无改动则完成 → 有改动则回到 3.3

---

## 4. 技能池扫描重构

### 测试用例设计
- [ ] 4.1 测试场景：
  - 全局扫描：遍历所有 KnownAgents 的 SkillsDir，返回已收录/新发现
  - 项目扫描：遍历 agent 子目录名匹配项目路径，只返回新发现
  - SkillsDir 不存在的 agent 被跳过不报错
  - 已收录技能 `AlreadyInPool=true`，新发现 `false`
  - 重复技能名去重（同一技能出现在多个 agent 目录）
  - 边界：项目路径为空时走全局扫描
  - 边界：项目路径无任何 agent 子目录时返回空
  - 边界：所有 agent 目录都无可读权限
- [ ] 4.2 Review 覆盖率：≥ 99%，无法覆盖的用例说明原因

### 开发实现
- [ ] 4.3 `backend/pkg/waillib/app.go` 中实现 `scanGlobalPool()`：
  - 遍历所有 KnownAgents 的 SkillsDir
  - 检查 index 是否已收录 → `AlreadyInPool=true/false`
  - 解析 SKILL.md 获取 version
- [ ] 4.4 实现 `scanProjectPool(projectPath)`：
  - 遍历 KnownAgents 的 SkillsDir 子目录名，在 `projectPath` 下匹配
  - 只返回 index 中不存在的技能（`AlreadyInPool=false`）
- [ ] 4.5 `ScanPool()` 入口：空 projectPath → `scanGlobalPool()`，非空 → `scanProjectPool()`
- [ ] 4.6 `SkillsPoolPage.tsx` UI 调整：
  - 空项目 → 标题"技能池 - 全局"，分两栏展示"已收录"/"新发现"
  - 有项目 → 标题"技能池 - 项目名"，只展示"新发现"

### 测试验证
- [ ] 4.7 `go build ./...` 通过
- [ ] 4.8 `go test ./backend/pkg/waillib/... -cover` 通过，覆盖率 ≥ 99%
- [ ] 4.9 `cd frontend && npm run build` 通过

### Code Review & 循环
- [ ] 4.10 Code Review → 无改动则完成 → 有改动则回到 4.3

---

## 5. 安装页 Agent 选择面板

### 测试用例设计
- [ ] 5.1 测试场景：
  - AgentSelector 渲染分组复选框（同路径合并）
  - displayName/tooltipName 与 SettingsPage 一致
  - localStorage 正确存取选中 agent IDs
  - 挂载时自动恢复上次选择
  - "全选已检测"只勾选 green dot 组
  - "清除选择"清空所有勾选和 localStorage
  - 折叠面板展开/收起交互
  - 无 agent 选择时安装按钮禁用
  - 安装时选中 agent IDs 正确传递给 `installSkill()`
  - 边界：localStorage 无数据时默认不勾选任何 agent
  - 边界：localStorage 数据中 agent ID 已在系统中不存在
- [ ] 5.2 Review 覆盖率：≥ 99%，无法覆盖的用例说明原因

### 开发实现
- [ ] 5.3 新建 `frontend/src/components/AgentSelector.tsx`：
  - Props: `agents: AgentGroup[]`, `selected: string[]`, `onChange: (ids: string[]) => void`
  - 渲染分组复选框，同路径合并一行
  - displayName/tooltipName 展示
  - 已检测绿色圆点 ●，未检测灰色 ○
- [ ] 5.4 实现 localStorage 存储/读取：
  - key: `skillsmanager:lastAgentSelection`
  - 每次 onChange 自动写入
  - 挂载时自动读取恢复勾选
- [ ] 5.5 实现"全选已检测"和"清除选择"按钮
- [ ] 5.6 实现折叠面板：搜索到结果后展开，无结果时折叠
- [ ] 5.7 整合到 `MarketPage.tsx`：
  - 搜索到结果后，在结果表和安装按钮之间插入 AgentSelector
  - 安装时从 AgentSelector 获取选中的 agent IDs 传给 `installSkill()`
  - 无 agent 选择时禁止点击安装按钮

### 测试验证
- [ ] 5.8 `cd frontend && npx tsc --noEmit` 通过
- [ ] 5.9 `cd frontend && npm run build` 通过

### Code Review & 循环
- [ ] 5.10 Code Review → 无改动则完成 → 有改动则回到 5.3

---

## 6. 全量构建验证
- [ ] 6.1 `cd frontend && npm run build` 前端构建通过
- [ ] 6.2 `go build ./...` 后端构建通过
- [ ] 6.3 `go test ./... -cover` 全部测试通过，整体覆盖率 ≥ 99%
- [ ] 6.4 `~/go/bin/wails build -s` Wails 完整构建通过
- [ ] 6.5 构建产物可打开运行