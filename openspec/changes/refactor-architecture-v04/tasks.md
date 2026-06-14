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
- **Review 覆盖率**：测试用例覆盖率必须 ≥ 99%，未能覆盖的场景需记录原因
- **测试必须 100% 通过**：后端 `go test ./... -cover` 全部通过，前端 `npm run build` 通过
- **Code Review** 由 requesting-code-review 技能执行
- Review 后若无改动 → 该阶段完成
- Review 后若有改动 → 回到开发实现，再次循环

---

## 阶段 1: 后端配置模型扩展

### 测试用例设计
- [ ] 1.1 测试场景：
  - Config 加载时 PoolPath 为空 → 自动设为默认值 ~/.skill-pool
  - Config 加载时 MarketSources 为空 → 返回空列表而非 nil
  - 保存配置 → 文件写入正确（含 PoolPath 和 MarketSources）
  - MarketSource 结构体 JSON 序列化/反序列化正确
  - 旧 Config 加载（无 PoolPath/MarketSources 字段）→ 不报错，自动填充默认值

### 实现
- [ ] 1.2 `backend/pkg/models/models.go`: Config 新增 `PoolPath string`, `MarketSources []MarketSource`
- [ ] 1.3 `backend/pkg/models/models.go`: 新增 `MarketSource` 结构体
- [ ] 1.4 `backend/internal/operations/config.go`: `defaultConfig()` 设置 PoolPath = ~/.skill-pool
- [ ] 1.5 `backend/internal/operations/config.go`: `LoadConfig()` 兼容旧配置（PoolPath 空时填充默认值）
- [ ] 1.6 `backend/internal/operations/config.go`: 确保 SaveConfig 保存所有新字段
- [ ] 1.7 测试 + 代码 Review 循环

## 阶段 2: Agent 检测重写

### 测试用例设计
- [ ] 2.1 测试场景：
  - CLI agent (`codex`, `claude-code` 等) 未安装 → `AutoDetected = false`
  - CLI agent 已安装 → `AutoDetected = true`（通过 exec.LookPath）
  - IDE agent (`cursor`, `trae` 等) 目录存在 → `AutoDetected = true`
  - IDE agent 目录不存在 → `AutoDetected = false`
  - SkillsDir 始终返回（无论检测结果）
  - 所有 48 个 agent 都分配了正确的 DetectCmd 或 DetectPath
  - Agent 列表排序：detected 在前，undetected 在后

### 实现
- [ ] 2.2 `backend/internal/distribute/agent.go`: Agent 结构体新增 `DetectCmd`, `DetectPath` 字段
- [ ] 2.3 为所有 48 个 agent 配置 DetectCmd（CLI）或 DetectPath（IDE）
- [ ] 2.4 `backend/internal/distribute/agent.go`: 重写 DetectAgents() → 用 exec.LookPath / os.Stat 代替 SkillsDir 检查
- [ ] 2.5 更新 `backend/pkg/waillib/app.go` 中 ListAgents() 使用新检测逻辑
- [ ] 2.6 测试 + 代码 Review 循环

## 阶段 3: 后端扫描拆分 + 池读取

### 测试用例设计
- [ ] 3.1 测试场景：
  - `ListPool()` 读取 PoolPath 目录 → 返回所有含 SKILL.md 的子目录
  - PoolPath 不存在 → 返回空列表而非报错
  - `ScanLocal(projectPath "")` 扫描所有 agent SkillsDir → 正确返回结果
  - `ScanLocal(projectPath "/tmp/xxx")` 组合全局 + 项目扫描
  - ScanLocal 结果标记 alreadyInPool（与池匹配）
  - ScanLocal 不修改池内容（只读扫描）
  - 池中没有 index 时 alsoInPool 全为 false

### 实现
- [ ] 3.2 `backend/pkg/waillib/app.go`: 新增 `ListPool() []DiscoveredSkill` 方法（读取 PoolPath）
- [ ] 3.3 `backend/pkg/waillib/app.go`: 重命名 `ScanPool()` → `ScanLocal(projectPath string)`
- [ ] 3.4 `backend/pkg/waillib/app.go`: ScanLocal 结果与池交叉匹配 → alreadyInPool 标记
- [ ] 3.5 `backend/pkg/waillib/app.go`: 在 GetConfig 中返回 PoolPath，新增 `SaveConfig(cfg) error`
- [ ] 3.6 `frontend/src/bridge.ts`: 新增 `listPool()`, `saveConfig(cfg)`, `scanLocal(path?)` API
- [ ] 3.7 测试 + 代码 Review 循环

## 阶段 4: Settings 页面增强

### 测试用例设计
- [ ] 4.1 测试场景：
  - Settings 页面加载 → 读取 config 显示当前 PoolPath
  - 修改 PoolPath 并保存 → 调用 saveConfig 持久化
  - 添加市场来源 → 表单验证 Name/URL/Type 必填
  - 删除市场来源 → 从列表中移除并保存
  - 启用/禁用市场来源 → toggle 切换并保存
  - 空列表时显示"暂无市场来源"引导提示
  - Type 为 "pool" 时 URL 必须为本地路径
  - Type 为 "github" 时 URL 必须为 github.com/owner/repo 格式

### 实现
- [ ] 4.2 `frontend/src/types.ts`: 新增 `MarketSource` 类型
- [ ] 4.3 `frontend/src/pages/SettingsPage.tsx`: 添加 PoolPath 配置卡片（输入框 + 保存按钮）
- [ ] 4.4 `frontend/src/pages/SettingsPage.tsx`: 添加市场来源管理卡片（列表 + 添加表单）
- [ ] 4.5 TypeScript 检查 + build 通过
- [ ] 4.6 代码 Review 循环

## 阶段 5: SkillsPoolPage 重构

### 测试用例设计
- [ ] 5.1 测试场景：
  - 页面加载 → 自动调用 listPool() 显示池内容
  - 池为空 → 显示"池为空，请先添加技能或扫描本机"引导
  - "本机扫描"按钮 → 弹出扫描对话框（可选项目路径）
  - 扫描完成后 → 分"已收录/未收录"两列显示
  - "导入到池"按钮（未收录技能上）→ 复制到 PoolPath
  - 导入后刷新池显示
  - 无 pool 路径配置 → 提示"请在设置中配置技能池路径"
  - PoolPath 目录不存在 → 自动创建提示

### 实现
- [ ] 5.2 `frontend/src/pages/SkillsPoolPage.tsx`: 重写为从 config 读取 PoolPath
- [ ] 5.3 `frontend/src/pages/SkillsPoolPage.tsx`: 添加自动列表池内容
- [ ] 5.4 `frontend/src/pages/SkillsPoolPage.tsx`: 添加"本机扫描"功能+已收录/未收录展示
- [ ] 5.5 `frontend/src/pages/SkillsPoolPage.tsx`: 添加"导入到池"按钮
- [ ] 5.6 TypeScript 检查 + build 通过
- [ ] 5.7 代码 Review 循环

## 阶段 6: MarketPage 重构

### 测试用例设计
- [ ] 6.1 测试场景：
  - 页面加载 → 从 config 读取 MarketSources
  - 无市场来源 → 显示"请在设置中添加市场来源"
  - 有来源 → 每个来源显示为一个搜索卡片（名称 + 类型标签 + 搜索按钮）
  - "搜索所有" → 按优先级搜索所有已启用来源
  - 结果按来源分组显示
  - 池来源搜索 → 直接读取 PoolPath 匹配
  - 安装按钮传递目标 agent 选择
  - 搜索加载状态和空结果处理

### 实现
- [ ] 6.2 `frontend/src/pages/MarketPage.tsx`: 重写为使用 config.MarketSources
- [ ] 6.3 每个来源渲染为搜索卡片
- [ ] 6.4 "搜索所有"按钮 → 池优先遍历
- [ ] 6.5 结果展示（按来源分组）
- [ ] 6.6 TypeScript 检查 + build 通过
- [ ] 6.7 代码 Review 循环

## 阶段 7: 全量构建验证

- [ ] 7.1 `go build ./...` 通过
- [ ] 7.2 `go test ./... -cover` 全部通过
- [ ] 7.3 `cd frontend && npm run build` 通过
- [ ] 7.4 `wails build` 通过