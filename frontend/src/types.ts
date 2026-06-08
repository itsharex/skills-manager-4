export type SourceType = "github" | "local" | "npm" | "registry";

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
}

export interface Agent {
  name: string;
  skill_location: string;
  global_location: string;
  installed: boolean;
  detected: boolean;
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
