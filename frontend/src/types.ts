export interface ListedSkill {
  name: string;
  agentIds: string[];
  agentNames: string[];
  paths: string[];
  storePath?: string; // pool storage path (e.g. ~/.skill-pool/<name>/)
  latest: string;
  versions: string[] | null;
  description: string;
  inPool: boolean;
}

export interface AgentInfo {
  id: string;
  name: string;
  path: string;
  skillsDir: string;
  projectSkillsSubdir: string;
  detected: boolean;
}

export interface Config {
  pool_path: string;
  install_mode: string;
  auto_fallback: boolean;
  default_agents: string[];
  market_sources: MarketSource[];
  link_targets: LinkTarget[];
  repositories: RepoSource[];
  cache_ttl: number;
  github_token: string;
}

export interface MarketSource {
  name: string;
  url: string;
  type: string;
  enabled: boolean;
  branch?: string;
}

export interface LinkTarget {
  id: string;
  path: string;
  enabled: boolean;
}

export interface RepoSource {
  name: string;
  url: string;
  type: string;
  enabled: boolean;
}

export interface ResolvedSkill {
  name: string;
  namespace: string;
  version: string;
  localPath: string;
}

export interface InstallResult {
  name: string;
  namespace: string;
  version: string;
  storePath: string;
  synced: boolean;
  syncMode: string;
  error: string;
}

export interface HealthReport {
  pool_path: string;
  checks: HealthCheck[];
  all_pass: boolean;
}

export interface HealthCheck {
  name: string;
  status: string;
  message: string;
}

export interface SkillStats {
  total_skills: number;
  total_versions: number;
  total_namespaces: number;
  total_agents: number;
  installed_skills: number;
  disk_usage_bytes: number;
}

export interface DiscoveredSkill {
  name: string;
  namespace: string;
  version: string;
  path: string;
  agentId?: string;
  agentName?: string;
  alreadyInPool: boolean;
}

export interface AgentGroup {
  path: string;
  agents: { id: string; name: string }[];
  detected: boolean;
  displayName: string;
  tooltipName: string;
}

// Market search result types
export interface MarketSearchResult {
  sourceName: string;
  sourceType: string; // "pool" | "clawhub" | "skillssh" | "github" | "registry"
  skills: MarketSearchSkill[];
  error?: string;
}

export interface OpLog {
  timestamp: string;
  operation: string;
  target: string;
  detail: string;
  source: string;
  storePath: string;
  agents: string;
  success: boolean;
  error?: string;
}

export interface MarketSearchSkill {
  name: string;
  namespace: string;
  version: string;
  description: string;
  source: string; // owner/repo for GitHub, owner/slug for ClawHub, owner/repo/slug for skills.sh
  localPath?: string; // local filesystem path (for pool skills)
  installs?: number;
}
