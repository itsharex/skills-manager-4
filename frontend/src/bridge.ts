// Bridge layer: wraps Wails-generated bindings with error handling.
import type { ListedSkill, AgentInfo, Config, ResolvedSkill, InstallResult, HealthReport, SkillStats, DiscoveredSkill, MarketSearchResult, OpLog } from "./types";

let WailsApp: any = null;

async function ensureWailsApp(): Promise<any> {
  if (WailsApp) return WailsApp;
  WailsApp = await import("../wailsjs/go/waillib/App");
  return WailsApp;
}

function hasWails(): boolean {
  return typeof window !== "undefined" && !!(window as any).go?.waillib?.App;
}

export async function getConfig(): Promise<Config> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.GetConfig() as Config;
}

export async function listSkills(): Promise<ListedSkill[]> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.ListSkills() as ListedSkill[];
}

export async function listAgents(): Promise<AgentInfo[]> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.ListAgents() as AgentInfo[];
}

export async function installSkill(source: string, opts: any): Promise<InstallResult[]> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.Install(source, opts) as InstallResult[];
}

export async function uninstallSkill(name: string, version: string): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  await app.Uninstall(name, version);
}

export async function searchSkills(source: string): Promise<ResolvedSkill[]> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.Search(source) as ResolvedSkill[];
}

export async function getStats(): Promise<SkillStats> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.GetStats() as SkillStats;
}

export async function runDoctor(): Promise<HealthReport> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.RunDoctor() as HealthReport;
}

export async function scanLocal(projectPath?: string): Promise<DiscoveredSkill[]> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.ScanLocal(projectPath || "") as DiscoveredSkill[];
}

export async function listPool(): Promise<DiscoveredSkill[]> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.ListPool() as DiscoveredSkill[];
}

export async function saveConfig(cfg: any): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  await app.SaveConfig(cfg);
}

export async function importToPool(sourcePath: string): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  await app.ImportToPool(sourcePath);
}

export async function deleteSkill(skillPath: string): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  await app.DeleteSkill(skillPath);
}

export async function deleteSkillFromPool(skillName: string): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return app.DeleteSkillFromPool(skillName);
}

export async function deleteSkillFromAgent(skillPath: string): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return app.DeleteSkillFromAgent(skillPath);
}

export async function deleteSkillFromProject(projectSkillPath: string): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return app.DeleteSkillFromProject(projectSkillPath);
}

export async function archiveToPool(sourcePath: string): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  await app.ArchiveToPool(sourcePath);
}

export async function openDirectoryDialog(title: string): Promise<string> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.OpenDirectoryDialog(title);
}

export async function openFileDialog(title: string): Promise<string> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.OpenFileDialog(title) as string;
}

export async function installToAgent(skillPath: string, agentSkillsDir: string, overwrite: boolean = false): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  await app.InstallToAgent(skillPath, agentSkillsDir, overwrite);
}

export async function openDirectory(dirPath: string): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  await app.OpenDirectory(dirPath);
}

export async function openURL(url: string): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  await app.OpenURL(url);
}

export async function installToProject(skillPath: string, projectPath: string, overwrite: boolean = false): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  await app.InstallToProject(skillPath, projectPath, overwrite);
}

export async function installToProjectForAgent(skillPath: string, projectPath: string, agentID: string, overwrite: boolean = false): Promise<void> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  await app.InstallToProjectForAgent(skillPath, projectPath, agentID, overwrite);
}

export async function searchMarket(keyword: string): Promise<MarketSearchResult[]> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.SearchMarket(keyword) as MarketSearchResult[];
}

export async function installMarketSkill(skill: any, agentIDs: string[]): Promise<any[]> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.InstallMarketSkill(skill, agentIDs) as any[];
}

export async function getOpLogs(n: number): Promise<OpLog[]> {
  if (!hasWails()) throw new Error("Wails backend not available");
  const app = await ensureWailsApp();
  return await app.GetOpLogs(n) as OpLog[];
}
