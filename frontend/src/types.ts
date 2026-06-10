export type SourceType = "github" | "local" | "npm" | "registry" | "clawhub";

export interface SkillSource {
  type: SourceType;
  url?: string;
  ref?: string;
  path?: string;
  command?: string;
  sub_path?: string;
}

export interface SkillVersion {
  version: string;
  installed_at: string;
  path: string;
  agents: string[];
}

export interface Skill {
  name: string;
  description: string;
  source: SkillSource;
  versions: Record<string, SkillVersion>;
  latest_version: string;
  user_tags?: string[];
}

// 技能安装状态："installed"（已安装到所有 Agent）| "partially_installed"（部分 Agent 安装）| "not_installed"（未安装）| "project_only"（仅项目目录有）
export type SkillInstallStatus =
  | "installed"
  | "partially_installed"
  | "not_installed"
  | "project_only";

// 带状态的技能（前端主视图使用）
export interface SkillWithStatus {
  name: string;
  description: string;
  tags?: string[];
  latestVersion: string;
  installStatus: SkillInstallStatus;
  installedAgents?: string[];
  totalAgents: number;
  source: SkillSource;
  sizeBytes?: number;
}

export interface ProjectSkill {
  name: string;
  description: string;
  path: string;
  isSymlink: boolean;
  symlinkTarget?: string;
  inLibrary: boolean;
  version?: string;
  tags?: string[];
  sizeBytes?: number;
}

export interface TagUsage {
  tag: string;
  count: number;
}

export interface Agent {
  name: string;
  skill_location: string;          // 项目级技能目录
  global_location: string;         // 全局技能目录
  installed: boolean;
  detected: boolean;
  supports_project: boolean;       // 是否支持项目级安装
  global_directory_key: string;    // 全局目录标识（共享检测）
  project_directory_key: string;   // 项目级目录标识（共享检测）
}

// 按目录分组的 Agent 集合
export interface AgentGroup {
  id: string;                       // 目录标识
  directory: string;                // 目录路径
  agentIds: string[];               // 该目录下的所有 Agent
  agentNames: string[];             // Agent 名称列表
  detectedIds: string[];            // 已检测的 Agent
  scope: "global" | "project";      // 作用域
  sharedRisk: boolean;              // 是否有共享目录风险（>=2 个 Agent）
}

export interface SkillspoolConfig {
  root: string;
}

export interface Config {
  skillspool: SkillspoolConfig;
  agents: Record<string, Agent>;
}

export interface InstallRequest {
  source: string;
  sub_path?: string;
  version?: string;
  ref?: string;
  agents?: string[];
  scope?: "global" | "project";
  project_dir?: string;
}

export interface InstallLink {
  agent_id: string;
  path: string;
  success: boolean;
  error?: string;
}

export interface InstallResult {
  skill_name: string;
  version: string;
  source: SkillSource;
  agents: Record<string, InstallLink>;
}

export interface MigrateResult {
  success: boolean;
  libraryPath?: string;
  symlinkCreated?: boolean;
  error?: string;
}

export interface BatchSyncRequest {
  skillNames: string[];
  agentIds: string[];
}

export interface BatchSyncItemResult {
  skillName: string;
  agentId: string;
  success: boolean;
  error?: string;
}

export interface BatchSyncResult {
  total: number;
  succeeded: number;
  failed: number;
  results: BatchSyncItemResult[];
}

// ---------- ClawHub 相关类型 ----------

export interface ClawHubSkill {
  owner: string;
  slug: string;
  name: string;
  description?: string;
  version?: string;
  author?: string;
  tags?: string[];
  downloads?: number;
  stars?: number;
  updatedAt?: string;
}

export interface RuntimeStatus {
  nodeInstalled: boolean;
  nodeVersion?: string;
  nodePath?: string;
  clawhubInstalled: boolean;
  clawhubVersion?: string;
  clawhubPath?: string;
  hasNpm: boolean;
  message?: string;
  registryReachable?: boolean;
  registryName?: string;
}

export interface RuntimeInstallReport {
  success: boolean;
  step?: string;
  message?: string;
  version?: string;
  path?: string;
}
