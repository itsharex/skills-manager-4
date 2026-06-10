import type {
  Skill,
  Agent,
  Config,
  InstallRequest,
  InstallResult,
  SkillWithStatus,
  ProjectSkill,
  TagUsage,
  MigrateResult,
  BatchSyncRequest,
  BatchSyncResult,
  ClawHubSkill,
  RuntimeStatus,
  AgentGroup,
} from "./types";
import * as WailsBindings from "../wailsjs/go/main/App";

// 声明 Window 类型
declare global {
  interface Window {
    go?: any;
  }
}

function hasWails(): boolean {
  return !!window.go?.main?.App;
}

// ================================================
// 基础查询 API
// ================================================

export async function listSkills(): Promise<Skill[]> {
  if (hasWails() && WailsBindings.ListSkills) {
    const skills = await WailsBindings.ListSkills();
    return skills as unknown as Skill[];
  }
  return mockListSkills();
}

export async function listAgents(): Promise<Record<string, Agent>> {
  if (hasWails() && WailsBindings.ListAgents) {
    return await WailsBindings.ListAgents();
  }
  return mockListAgents();
}

export async function listAgentGroups(scope: "global" | "project"): Promise<AgentGroup[]> {
  if (hasWails() && (WailsBindings as any).ListAgentGroups) {
    const groups = await (WailsBindings as any).ListAgentGroups(scope);
    return groups as unknown as AgentGroup[];
  }
  // mock 返回空列表给无 Wails 环境
  return [];
}

export async function getConfig(): Promise<Config> {
  if (hasWails() && WailsBindings.GetConfig) {
    return await WailsBindings.GetConfig();
  }
  return mockGetConfig();
}

export async function installSkill(req: InstallRequest): Promise<InstallResult[]> {
  if (hasWails() && WailsBindings.Install) {
    const wailsReq = { ...req, agents: req.agents || [] };
    const results = await WailsBindings.Install(wailsReq as any);
    return results as unknown as InstallResult[];
  }
  return mockInstall(req);
}

// 获取/设置技能池根目录
export async function getSkillspoolRoot(): Promise<string> {
  if (hasWails() && WailsBindings.GetSkillspoolRoot) {
    return await WailsBindings.GetSkillspoolRoot();
  }
  if (hasWails() && WailsBindings.SkillspoolRoot) {
    return await WailsBindings.SkillspoolRoot();
  }
  return "~/SkillsManager/skillspool";
}

export async function setSkillspoolRoot(root: string): Promise<any> {
  if (hasWails() && WailsBindings.SetSkillspoolRoot) {
    return await WailsBindings.SetSkillspoolRoot(root);
  }
  return { success: true, old_root: "", new_root: root, moved_files: 0, message: "mock: 模拟成功" };
}

// 选择本地目录
export async function selectDirectory(title: string): Promise<string> {
  if (hasWails() && WailsBindings.SelectDirectory) {
    return await WailsBindings.SelectDirectory(title);
  }
  return "";
}

export async function selectFile(title: string): Promise<string> {
  if (hasWails() && WailsBindings.SelectFile) {
    return await WailsBindings.SelectFile(title);
  }
  return "";
}

// ================================================
// 扩展 API：带状态的技能、标签、项目技能、迁移
// ================================================

export async function listSkillsWithStatus(): Promise<SkillWithStatus[]> {
  if (hasWails() && WailsBindings.ListSkillsWithStatus) {
    return (await WailsBindings.ListSkillsWithStatus()) as unknown as SkillWithStatus[];
  }
  return mockListSkillsWithStatus();
}

export async function addSkillTag(skillName: string, tag: string): Promise<boolean> {
  if (hasWails() && WailsBindings.AddSkillTag) {
    return await WailsBindings.AddSkillTag(skillName, tag);
  }
  return true;
}

export async function removeSkillTag(skillName: string, tag: string): Promise<boolean> {
  if (hasWails() && WailsBindings.RemoveSkillTag) {
    return await WailsBindings.RemoveSkillTag(skillName, tag);
  }
  return true;
}

export async function getSkillTags(skillName: string): Promise<string[]> {
  if (hasWails() && WailsBindings.GetSkillTags) {
    return await WailsBindings.GetSkillTags(skillName);
  }
  return [];
}

export async function getAllTags(): Promise<TagUsage[]> {
  if (hasWails() && WailsBindings.GetAllTags) {
    return await WailsBindings.GetAllTags();
  }
  return [];
}

export async function installSkillToAgent(skillName: string, agentID: string): Promise<boolean> {
  if (hasWails() && WailsBindings.InstallSkillToAgent) {
    try {
      const ok = await WailsBindings.InstallSkillToAgent(skillName, agentID);
      return !!ok;
    } catch (err) {
      console.error("installSkillToAgent error:", err);
      return false;
    }
  }
  return true;
}

export async function uninstallSkillFromAgent(skillName: string, agentID: string): Promise<boolean> {
  if (hasWails() && WailsBindings.UninstallSkillFromAgent) {
    try {
      const ok = await WailsBindings.UninstallSkillFromAgent(skillName, agentID);
      return !!ok;
    } catch (err) {
      console.error("uninstallSkillFromAgent error:", err);
      return false;
    }
  }
  return true;
}

// 扫描项目目录下的技能
export async function scanProjectSkills(projectPath: string): Promise<ProjectSkill[]> {
  if (hasWails() && WailsBindings.ScanProjectSkills) {
    return await WailsBindings.ScanProjectSkills(projectPath);
  }
  return mockScanProjectSkills(projectPath);
}

// 迁移项目技能到库
export async function migrateProjectSkillToLibrary(skillPath: string, projectPath: string): Promise<MigrateResult> {
  if (hasWails() && WailsBindings.MigrateProjectSkillToLibrary) {
    return await WailsBindings.MigrateProjectSkillToLibrary(skillPath, projectPath);
  }
  return { success: true, libraryPath: skillPath };
}

// 批量同步
export async function batchSyncSkills(req: BatchSyncRequest): Promise<BatchSyncResult> {
  if (hasWails() && WailsBindings.BatchSyncSkills) {
    return await WailsBindings.BatchSyncSkills(req);
  }
  return {
    total: req.skillNames.length * req.agentIds.length,
    succeeded: req.skillNames.length * req.agentIds.length,
    failed: 0,
    results: [],
  };
}

// ================================================
// ClawHub 市场 API
// ================================================

export async function runtimeStatus(): Promise<RuntimeStatus> {
  if (hasWails() && WailsBindings.RuntimeStatus) {
    return await WailsBindings.RuntimeStatus();
  }
  return {
    nodeInstalled: false,
    clawhubInstalled: false,
    hasNpm: false,
    message: "运行时检测不可用（仅 Wails 环境）",
  };
}

export async function ensureRuntime(): Promise<RuntimeStatus> {
  if (hasWails() && WailsBindings.EnsureRuntime) {
    const result = await WailsBindings.EnsureRuntime();
    // result 可能是 [status, error] 元组或直接对象
    if (Array.isArray(result)) {
      return (result[0] as RuntimeStatus) || { nodeInstalled: false, clawhubInstalled: false, hasNpm: false, message: "安装失败" };
    }
    return result as RuntimeStatus;
  }
  return { nodeInstalled: false, clawhubInstalled: false, hasNpm: false, message: "运行时检测不可用" };
}

export async function searchClawHub(keyword: string): Promise<ClawHubSkill[]> {
  if (hasWails() && WailsBindings.SearchClawHub) {
    const result = await WailsBindings.SearchClawHub(keyword);
    return (result as ClawHubSkill[]) || [];
  }
  return mockSearchClawHub(keyword);
}

export async function installFromClawHub(owner: string, slug: string, agentIds: string[]): Promise<InstallResult> {
  if (hasWails() && WailsBindings.InstallFromClawHub) {
    const result = await WailsBindings.InstallFromClawHub(owner, slug, agentIds || []);
    if (Array.isArray(result)) {
      return (result[0] as InstallResult) || (result as unknown as InstallResult);
    }
    return result as InstallResult;
  }
  return {
    skill_name: `${owner}/${slug}`,
    version: "1.0.0",
    source: { type: "clawhub", url: `https://clawhub.ai/${owner}/${slug}` },
    agents: {},
  };
}

// ================================================
// MOCK 数据（无 Wails 环境时用）
// ================================================

function mockListSkills(): Skill[] {
  return [
    {
      name: "frontend-design",
      description: "Create production-grade web interfaces with high design quality",
      source: { type: "local", path: "/path/to/frontend-design" },
      versions: {
        "1.0.0": {
          version: "1.0.0",
          installed_at: "2026-06-01T00:00:00Z",
          path: "frontend-design/1.0.0",
          agents: ["trae", "claude"],
        },
        "1.2.0": {
          version: "1.2.0",
          installed_at: "2026-06-08T00:00:00Z",
          path: "frontend-design/1.2.0",
          agents: ["trae", "claude", "cursor"],
        },
      },
      latest_version: "1.2.0",
    },
    {
      name: "webapp-testing",
      description: "Toolkit for building tests for web applications",
      source: { type: "github", url: "https://github.com/example/webapp-testing" },
      versions: {
        "0.5.0": {
          version: "0.5.0",
          installed_at: "2026-06-05T00:00:00Z",
          path: "webapp-testing/0.5.0",
          agents: ["trae"],
        },
      },
      latest_version: "0.5.0",
    },
    {
      name: "skill-creator",
      description: "Help users create new skills",
      source: { type: "local", path: "/path/to/skill-creator" },
      versions: {
        "1.0.0": {
          version: "1.0.0",
          installed_at: "2026-06-01T00:00:00Z",
          path: "skill-creator/1.0.0",
          agents: [],
        },
      },
      latest_version: "1.0.0",
    },
  ];
}

function mockListSkillsWithStatus(): SkillWithStatus[] {
  return [
    {
      name: "frontend-design",
      description: "Create production-grade web interfaces with high design quality",
      tags: ["design", "frontend"],
      latestVersion: "1.2.0",
      installStatus: "installed",
      installedAgents: ["trae", "claude", "cursor"],
      totalAgents: 3,
      source: { type: "local", path: "/path/to/frontend-design" },
    },
    {
      name: "webapp-testing",
      description: "Toolkit for building tests for web applications",
      tags: ["testing"],
      latestVersion: "0.5.0",
      installStatus: "partially_installed",
      installedAgents: ["trae"],
      totalAgents: 3,
      source: { type: "github", url: "https://github.com/example/webapp-testing" },
    },
    {
      name: "skill-creator",
      description: "Help users create new skills",
      tags: [],
      latestVersion: "1.0.0",
      installStatus: "not_installed",
      installedAgents: [],
      totalAgents: 3,
      source: { type: "local", path: "/path/to/skill-creator" },
    },
  ];
}

function mockListAgents(): Record<string, Agent> {
  return {
    trae: {
      name: "Trae",
      skill_location: ".trae/skills",
      global_location: "~/.trae-cn/skills",
      installed: true,
      detected: true,
      supports_project: true,
      global_directory_key: ".trae-cn/skills",
      project_directory_key: ".trae/skills",
    },
    claude: {
      name: "Claude Code",
      skill_location: ".claude/skills",
      global_location: "~/.claude/skills",
      installed: true,
      detected: true,
      supports_project: true,
      global_directory_key: ".claude/skills",
      project_directory_key: ".claude/skills",
    },
    cursor: {
      name: "Cursor",
      skill_location: ".agents/skills",
      global_location: "~/.cursor/skills",
      installed: false,
      detected: false,
      supports_project: true,
      global_directory_key: ".cursor/skills",
      project_directory_key: ".agents/skills",
    },
  };
}

function mockGetConfig(): Config {
  return {
    skillspool: { root: "~/Library/Application Support/SkillsManager/skillspool" },
    agents: mockListAgents(),
  };
}

function mockInstall(req: InstallRequest): Promise<InstallResult[]> {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve([
        {
          skill_name: "demo-skill",
          version: "1.0.0",
          source: { type: "local", path: req.source },
          agents: (req.agents || ["trae"]).reduce((acc, id) => {
            acc[id] = {
              agent_id: id,
              path: `/path/to/${id}/skills/demo-skill`,
              success: true,
            };
            return acc;
          }, {} as Record<string, InstallResult["agents"][string]>),
        },
      ]);
    }, 600);
  });
}

function mockScanProjectSkills(projectPath: string): ProjectSkill[] {
  return [
    {
      name: "project-custom-skill",
      description: "Custom skill only in this project",
      path: `${projectPath}/.trae/skills/project-custom-skill`,
      isSymlink: false,
      inLibrary: false,
      version: "0.1.0",
      tags: ["project"],
    },
    {
      name: "frontend-design",
      description: "Create production-grade web interfaces",
      path: `${projectPath}/.trae/skills/frontend-design`,
      isSymlink: false,
      inLibrary: true,
      version: "1.2.0",
      tags: ["design", "frontend"],
    },
  ];
}

function mockSearchClawHub(keyword: string): ClawHubSkill[] {
  const demo: ClawHubSkill[] = [
    {
      owner: "openclaw",
      slug: "python-developer",
      name: "openclaw/python-developer",
      description: "Python 开发相关的命令与工作流集合",
      version: "1.0.0",
      tags: ["python", "dev"],
      downloads: 12000,
      stars: 312,
      updatedAt: "2026-05-15",
    },
    {
      owner: "community",
      slug: "react-builder",
      name: "community/react-builder",
      description: "快速构建 React 组件的技能包",
      version: "0.8.2",
      tags: ["react", "frontend"],
      downloads: 9800,
      stars: 251,
      updatedAt: "2026-05-20",
    },
    {
      owner: "cloudops",
      slug: "aws-cli",
      name: "cloudops/aws-cli",
      description: "AWS CLI 操作相关的辅助指令",
      version: "2.0.0",
      tags: ["aws", "cloud"],
      downloads: 7500,
      stars: 188,
      updatedAt: "2026-04-10",
    },
  ];
  const k = keyword.trim().toLowerCase();
  if (!k) return demo;
  return demo.filter((s) =>
    (s.name + " " + (s.description || "") + " " + (s.tags || []).join(" "))
      .toLowerCase()
      .includes(k)
  );
}
