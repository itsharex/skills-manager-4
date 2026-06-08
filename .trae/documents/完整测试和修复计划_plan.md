
# Skills Manager 完整测试和修复计划

## 1. 问题分析

### 已发现的问题：

#### 1.1 Config 类型绑定问题
- **位置**: `backend/pkg/models/models.go` 的 `Config` 结构体
- **问题**: `Skillspool` 字段使用了匿名结构体，导致 Wails 绑定生成器无法正确识别类型
- **影响**: 在 `frontend/wailsjs/go/models.ts` 中，`skillspool` 类型被标记为 `any`

#### 1.2 app.go 类型冗余问题
- **位置**: `app.go`
- **问题**: `app.go` 中重复定义了 `InstallRequest`、`InstallResult`、`InstallLink` 类型，这些类型已在 `backend/pkg/api/api.go` 中定义
- **影响**: 类型重复，可能导致混乱

#### 1.3 Install 方法返回类型不匹配
- **位置**: `app.go` 的 `Install` 方法
- **问题**: 返回类型是 `[]skillsapi.InstallResult`，但前端类型期望 `InstallResult` 数组
- **需要确认**: 确保类型兼容性

#### 1.4 wails.json 配置不完整
- **位置**: `wails.json`
- **问题**: 当前配置过于简单，缺少前端构建配置
- **影响**: 无法使用 `wails dev` 自动构建前端

#### 1.5 bridge.ts 未使用自动生成绑定
- **位置**: `frontend/src/bridge.ts`
- **问题**: 当前 bridge.ts 手动定义了绑定调用，没有使用 `frontend/wailsjs/go/main/App.js` 中自动生成的绑定函数

---

## 2. 修复步骤

### 步骤 1: 修复 Config 结构体类型
- 为 `Skillspool` 定义具名结构体
- 修改 `backend/pkg/models/models.go`

### 步骤 2: 清理 app.go 重复类型
- 删除 `app.go` 中重复定义的类型
- 直接使用 `skillsapi` 包中的类型
- 简化代码，避免冗余

### 步骤 3: 完善 wails.json 配置
- 添加完整的 Wails 配置
- 配置前端安装、构建、开发命令
- 配置正确的前端输出目录

### 步骤 4: 更新 bridge.ts 使用自动生成的绑定
- 修改 `frontend/src/bridge.ts`
- 使用 `frontend/wailsjs/go/main/App.js` 中的函数
- 保持 mock 功能的同时使用真实绑定

### 步骤 5: 完整测试流程
- 重新生成 Wails 绑定
- 重新构建前端
- 使用 `wails dev` 进行开发模式测试
- 测试所有功能页面
- 测试 CLI 命令
- 测试最终应用包

---

## 3. 具体修改文件

### 需要修改的文件：

1. **`backend/pkg/models/models.go`**
   - 新增具名的 `SkillspoolConfig` 结构体
   - 替换 `Config` 中的匿名 `Skillspool` 字段

2. **`app.go`**
   - 删除重复的类型定义
   - 使用 `skillsapi` 包中的类型
   - 更新 `Install` 方法实现

3. **`wails.json`**
   - 添加完整配置项
   - 配置正确的前端路径

4. **`frontend/src/bridge.ts`**
   - 使用 Wails 自动生成的绑定
   - 保持 mock 功能完整性

---

## 4. 测试策略

### 4.1 单元测试
- 验证 Config 类型修复
- 验证类型兼容性

### 4.2 集成测试
- 测试完整的开发模式流程 (`wails dev`)
- 测试前端与后端通信
- 测试技能安装功能

### 4.3 完整构建测试
- 测试 `wails build` 完整流程
- 验证应用包功能

### 4.4 CLI 验证
- 验证 CLI 命令功能完整性

---

## 5. 预期结果

修复后应该能够：
- ✅ 使用 `wails dev` 正常启动开发模式
- ✅ GUI 应用完整可用（安装技能、查看已安装技能、查看 Agent 配置）
- ✅ 类型绑定正确，无 TypeScript 错误
- ✅ CLI 命令完整可用
- ✅ 完整构建包功能正常

