import type { Skill, Agent, Config, InstallRequest, InstallResult } from "./types";
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

export async function skillspoolRoot(): Promise<string> {
  if (hasWails() && WailsBindings.SkillspoolRoot) {
    return await WailsBindings.SkillspoolRoot();
  }
  return "~/Library/Application Support/SkillsManager/skillspool";
}

// ----------------- MOCKS (for preview without Wails) -----------------
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
      name: "@myorg/scoped-skill",
      description: "A scoped skill example",
      source: { type: "github", url: "https://github.com/myorg/skills" },
      versions: {
        "0.1.0": {
          version: "0.1.0",
          installed_at: "2026-06-08T00:00:00Z",
          path: "@myorg/scoped-skill/0.1.0",
          agents: ["trae"],
        },
      },
      latest_version: "0.1.0",
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
    },
    claude: {
      name: "Claude Code",
      skill_location: ".claude/skills",
      global_location: "~/.claude/skills",
      installed: true,
      detected: true,
    },
    cursor: {
      name: "Cursor",
      skill_location: ".cursor/skills",
      global_location: "~/.cursor/skills",
      installed: false,
      detected: false,
    },
    windsurf: {
      name: "Windsurf",
      skill_location: ".windsurf/skills",
      global_location: "~/.windsurf/skills",
      installed: false,
      detected: false,
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
          skill_name: "install-demo-skill",
          version: "1.0.0",
          source: { type: "local", path: req.source },
          agents: (req.agents || ["trae"]).reduce((acc, id) => {
            acc[id] = {
              agent_id: id,
              path: `/path/to/${id}/skills/install-demo-skill`,
              success: true,
            };
            return acc;
          }, {} as Record<string, InstallResult["agents"][string]>),
        },
      ]);
    }, 800);
  });
}
