# Skills Manager v0.3 打磨设计

## 目标

打磨 v0.2 重构后的 Skills Manager，解决以下问题：
1. 移除前端 mock/demo 数据
2. 扩展 agent 列表至 40+
3. 实现 Agent 分组（同路径合并）+ 已检测排序优先
4. 完善技能池扫描逻辑

---

## 问题 1：移除 mock/demo 数据

### 现状

`frontend/src/bridge.ts` 中所有 API 函数都有 mock fallback：

```typescript
// 无 Wails 时返回假数据
export async function listAgents(): Promise<AgentInfo[]> {
  if (hasWails()) return await wailsBackend.ListAgents();
  return mockAgents(); // ← 需移除
}
```

当 Wails 绑定失败时，用户看到的是假数据而非真实错误。

### 方案

删除所有 mock 函数（`mockConfig`, `mockSkills`, `mockAgents`, `mockSearch`, `mockInstall`），改为直接抛出错误：

```typescript
export async function listAgents(): Promise<AgentInfo[]> {
  if (!hasWails()) throw new Error("Wails backend not available");
  return await wailsBackend.ListAgents();
}
```

### 文件变更

- `frontend/src/bridge.ts`：删除所有 mock 函数，API 函数改为抛错

---

## 问题 2：扩展 agent 列表至 40+

### 现状

`backend/internal/distribute/agent.go` 中 `KnownAgents()` 共 30 个 agent，但缺少部分流行工具。

### 方案

补充以下 agent（按流行度）：

| ID | Name | SkillsDir |
|----|------|-----------|
| VS Code | VS Code (Copilot Chat) | ~/.vscode/extensions/github.copilot-chat-<id>/skills |
| vscode-copilot | VS Code Copilot | ~/.vscode/extensions/github.copilot-<id>/skills |
| copilot-chat | Copilot Chat | ~/.vscode/extensions/github.copilot-chat-<id>/skills |
| zed | Zed AI | ~/.config/zed/settings.json (检测) |
| zed-dev | Zed Developer | ~/.config/zed/skills |
| jupyter | Jupyter AI | ~/.jupyter/ai/skills |
| goose | Google Goose | ~/.goose/skills |
| devin | Cognition Devin | ~/.devin/skills |
| aic | Amazon CodeWhisperer | ~/.aws/amazon-q/skills |
| bolt | Bolt (Supermaven) | ~/.bolt/skills |

最终目标：40+ agent。

### 文件变更

- `backend/internal/distribute/agent.go`：补充 10 个 agent

---

## 问题 3：Agent 分组 + 已检测排序优先

### 现状

`ListAgents()` 返回扁平的 `AgentInfo[]`，未分组。

### 目标行为

1. **同路径合并**：共享 `SkillsDir` 的 agent 分为一组
2. **显示名字**：最短名字优先，如 "Trae, Trae CN" → "Trae"
3. **无检测时**："名字 + 等"，如 "Trae 等"（tooltip 展示全部）
4. **已检测排序**：所有已检测的组排前面，未检测的组排后面

### 类型设计

```typescript
// frontend/src/types.ts
interface AgentGroup {
  path: string;         // SkillsDir 路径（分组 key）
  agents: { id: string; name: string }[];  // 同路径的 agent 列表
  detected: boolean;     // 该组是否至少有一个 agent 被检测到
  displayName: string;   // 用于 UI 显示的名字（最短 + 截断逻辑）
  tooltipName: string;  // 完整 tooltip 文本
}
```

### 前端逻辑

```typescript
function buildAgentGroups(detected: Agent[]): AgentGroup[] {
  // 1. 按 SkillsDir 分组
  const pathMap = new Map<string, Agent[]>();
  for (const ag of detected) {
    if (!pathMap.has(ag.SkillsDir)) pathMap.set(ag.SkillsDir, []);
    pathMap.get(ag.SkillsDir)!.push(ag);
  }
  // 2. 生成 displayName/tooltipName
  for (const [path, agents] of pathMap) {
    const names = agents.map(a => a.Name);
    // 按名字长度升序，最短的作为 displayName
    names.sort((a, b) => a.length - b.length);
    group.displayName = names.length === 1 ? names[0]
      : detected ? names[0] : `${names[0]} 等`;
    group.tooltipName = names.join(", ");
  }
  // 3. detected 组排前面
  return sorted([...pathMap.values()]);
}
```

### 文件变更

- `frontend/src/types.ts`：新增 `AgentGroup` 类型
- `frontend/src/App.tsx` 或专用 hook：实现分组逻辑
- 安装技能 UI（MarketPage 或新组件）：接收 `AgentGroup[]` 渲染复选框

### 后端改动

`waillib.App.ListAgents()` 保持返回扁平时 `AgentInfo[]`，前端做分组聚合。

---

## 问题 4：技能池扫描逻辑

### 现状

`ScanPool()` 部分实现了两种模式，但逻辑不完整。

### 目标行为

**无项目路径时（全局池）：**
- 扫描所有 40+ agent 的全局 SkillsDir
- 结果分两组展示：
  - **已收录**：技能已在 index 中，显示来源 agent
  - **新发现**：SkillsDir 中有但 index 中没有的技能

**有项目路径时（项目扫描）：**
- 扫描该项目目录下各 agent 子目录（`.claude/skills`、`./.cursor/skills` 等）
- 只显示**未收录**到池中的技能（列"新发现"）
- 不显示已收录的技能

### 后端改动

`waillib.App.ScanPool(projectPath string)`：

```go
func (a *App) ScanPool(projectPath string) []DiscoveredSkill {
  if projectPath == "" {
    return a.scanGlobalPool()   // 全局扫描
  }
  return a.scanProjectPool(projectPath)  // 项目扫描
}

func (a *App) scanGlobalPool() []DiscoveredSkill {
  // 遍历所有 KnownAgents 的 SkillsDir
  // 对每个技能目录：
  //   - 检查 index 是否已收录 → AlreadyInPool=true/false
  //   - 解析 SKILL.md 获取 version
  //   - 返回 DiscoveredSkill
}

func (a *App) scanProjectPool(projectPath string) []DiscoveredSkill {
  // 遍历所有 KnownAgents 的 skills 子目录名
  // 扫描 projectPath 下所有 "{subdir}/SKILL.md"
  // 只返回 index 中不存在的技能 → AlreadyInPool=false
}
```

### 前端改动

`SkillsPoolPage.tsx`：
- 空项目路径时：标题改为"技能池 - 全局"，分"已收录"/"新发现"两组
- 有项目路径时：标题改为"技能池 - 项目名"，只显示"新发现"
- 不再有 poolSkills/newSkills 的双重分类逻辑（由后端统一归类）

### 文件变更

- `backend/pkg/waillib/app.go`：`ScanPool` 拆分 + 增强逻辑
- `frontend/src/pages/SkillsPoolPage.tsx`：UI 展示调整

---

## 问题 5：安装页新增 Agent 选择面板（记住选择）

### 现状

MarketPage 搜索到技能后，直接调用 `installSkill(source, {})` 安装，无 agent 选择环节。

### 目标行为

**搜索到技能后，在安装按钮前插入 agent 选择面板：**

1. **分组展示**：使用与 SettingsPage 相同的 AgentGroup 逻辑，同路径 agent 合并为一条
2. **记住选择**：用户勾选后，将选中的 agent IDs 保存到 `localStorage`（key: `lastAgentSelection`）
3. **预勾选**：每次打开安装页时，自动从 localStorage 读取上次选择。无记录时不勾选
4. **全选/清除**：提供"全选已检测"和"清除选择"快捷按钮

### UI 布局

```
┌─────────────────────────────────────────┐
│  搜索技能                                │
│  [输入框] [搜索按钮]                      │
├─────────────────────────────────────────┤
│  ▼ 安装到 Agent（上次选择: Trae, Cursor） │ ← 可折叠面板
│  ┌─────────────────────────────────────┐│
│  │ [全选已检测] [清除选择]              ││
│  │                                     ││
│  │ ✅ 📍 ~/.trae-cn/skills → Trae     ││ ← 已检测组排前面
│  │ ✅ 📍 ~/.cursor/skills → Cursor    ││
│  │ ☐ 📍 ~/.claude/skills → Claude Code││
│  │    ...                              ││
│  │                                     ││
│  │ 📍 ~/.windsurf/skills → Windsurf   ││ ← 未检测组，灰色
│  │ 📍 ~/.devin/skills → Devin         ││
│  └─────────────────────────────────────┘│
│  [安装技能]                              │
└─────────────────────────────────────────┘
```

### 交互细节

- 已检测的组前面显示绿色圆点 ●，未检测的组显示灰色 ○
- 同路径 agent 悬停显示 tooltip（如 "Trae, Trae CN, Cursor"）
- "全选已检测"只勾选绿色圆点的组
- 折叠面板可展开/收起，减少页面跳动

### localStorage 格式

```typescript
const STORAGE_KEY = "skillsmanager:lastAgentSelection";
// 存储格式: string[] — agent IDs 数组
// 示例: ["trae", "cursor", "claude-code"]
```

### 文件变更

- `frontend/src/pages/MarketPage.tsx`：新增 AgentSelector 组件 + 记住选择逻辑
- `frontend/src/components/AgentSelector.tsx`：可复用的 agent 选择面板组件

---

## 实现顺序

1. 清理 mock 数据（`bridge.ts`）
2. 扩展 agent 列表（`agent.go`）
3. 前端 AgentGroup `displayName`/`tooltipName` 增强（`types.ts` + `SettingsPage.tsx`）
4. 后 + 前端 `ScanPool` 重构（`app.go` + `SkillsPoolPage.tsx`）
5. MarketPage 安装页新增 agent 选择面板（AgentSelector 组件 + 记住选择）
6. 全量构建验证

---

## 验收标准

- [ ] `bridge.ts` 中无任何 mock 函数
- [ ] `KnownAgents()` 包含 40+ agent
- [ ] Agent 分组显示中已检测的组排前面
- [ ] 同路径 agent 合并显示（最短名 + tooltip）
- [ ] 无 Agent 时也显示"名 + 等"截断
- [ ] 无项目时扫描全局 SkillsDir，分已收录/新发现两组
- [ ] 有项目时沿 agent 子目录名扫描，只显示新发现技能
- [ ] MarketPage 安装前展示 agent 选择面板
- [ ] agent 选择会记住上次勾选（localStorage）
- [ ] 前端构建通过，Wails 完整构建通过
