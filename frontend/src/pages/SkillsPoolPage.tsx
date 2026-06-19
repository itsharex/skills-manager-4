import { useState, useEffect } from "react";
import { RefreshCw, Search, Loader2, Puzzle, CheckCircle2, FolderOpen, Download, Archive, ExternalLink, Plus, Circle, Trash2, HelpCircle } from "lucide-react";
import { toast } from "sonner";
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "../components/ui/tabs";
import { scanLocal, getConfig, importToPool, archiveToPool, openDirectoryDialog, installToAgent, listAgents, openDirectory, installToProjectForAgent, deleteSkillFromPool, deleteSkillFromAgent, deleteSkillFromProject } from "../bridge";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "../components/ui/dialog";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "../components/ui/alert-dialog";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "../components/ui/tooltip";
import { Label } from "../components/ui/label";
import { Input } from "../components/ui/input";
import type { ListedSkill, DiscoveredSkill, AgentInfo } from "../types";

interface Props {
  skills: ListedSkill[];
  onSelect: (skill: ListedSkill) => void;
  onRefresh: () => void;
  loading?: boolean;
}

export default function SkillsPoolPage({ skills, onSelect, onRefresh, loading }: Props) {
  // Pool state
  const [poolPath, setPoolPath] = useState("");
  const [loadingPool, setLoadingPool] = useState(true);

  // Scan state
  const [projectPath, setProjectPath] = useState("");
  const [scanning, setScanning] = useState(false);
  const [scanResults, setScanResults] = useState<DiscoveredSkill[]>([]);
  const [scanError, setScanError] = useState<string | null>(null);

  // Action state
  const [importing, setImporting] = useState(false);

  // Multi-select state for scan results (by skill name)
  const [selectedScanSkills, setSelectedScanSkills] = useState<Set<string>>(new Set());

  // Pool filter state
  const [poolKeyword, setPoolKeyword] = useState("");

  // Global skills filter state
  const [globalKeyword, setGlobalKeyword] = useState("");
  const [selectedGlobalSkills, setSelectedGlobalSkills] = useState<Set<string>>(new Set());

  // Multi-select state for pool skills (batch install)
  const [selectedPoolSkills, setSelectedPoolSkills] = useState<Set<string>>(new Set());

  // Install state
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [installing, setInstalling] = useState(false);
  const [installDialogOpen, setInstallDialogOpen] = useState(false);
  const [selectedSkillsForInstall, setSelectedSkillsForInstall] = useState<ListedSkill[]>([]);
  const [selectedAgents, setSelectedAgents] = useState<Set<string>>(new Set());
  const [installModeTab, setInstallModeTab] = useState("agent");
  const [projectInstallPath, setProjectInstallPath] = useState("");
  const [openingFolder, setOpeningFolder] = useState<string | null>(null);

  // Install result state
  interface InstallResultItem { skillName: string; agentName: string; success: boolean; error?: string }
  const [installResults, setInstallResults] = useState<InstallResultItem[]>([]);
  const [installResultOpen, setInstallResultOpen] = useState(false);

  // Delete state (three levels)
  const [deleting, setDeleting] = useState(false);
  type DeleteTarget = { level: "pool" | "agent" | "project"; skill: ListedSkill | DiscoveredSkill; path: string; paths?: string[] };
  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const cfg = await getConfig();
        setPoolPath(cfg.pool_path || "");
      } catch { /* ignore */ }
      finally { setLoadingPool(false); }
    }
    load();
    loadAgents();
  }, []);

  const loadAgents = async () => {
    try {
      const ag = await listAgents();
      setAgents(ag);
    } catch { /* ignore */ }
  };

  const handleInstall = async () => {
    if (selectedSkillsForInstall.length === 0 || selectedAgents.size === 0) return;
    setInstalling(true);
    const results: InstallResultItem[] = [];
    for (const skill of selectedSkillsForInstall) {
      for (const agentId of selectedAgents) {
        const agent = agents.find(a => a.id === agentId);
        if (!agent || !agent.skillsDir) continue;
        try {
          await installToAgent(skill.storePath || skill.paths[0], agent.skillsDir, true);
          results.push({ skillName: skill.name, agentName: agent.name, success: true });
        } catch (e: any) {
          const errMsg = typeof e === "string" ? e : (e?.message || e?.toString?.() || String(e));
          results.push({ skillName: skill.name, agentName: agent.name, success: false, error: errMsg });
        }
      }
    }
    const totalSuccess = results.filter(r => r.success).length;
    const totalFail = results.filter(r => !r.success).length;
    const skillCount = selectedSkillsForInstall.length;
    const agentCount = selectedAgents.size;

    if (totalFail === 0) {
      toast.success(`成功安装 ${skillCount} 个技能到 ${agentCount} 个智能体`);
    } else {
      toast.error(`安装完成：${totalSuccess} 成功，${totalFail} 失败`);
      setInstallResults(results);
      setInstallResultOpen(true);
    }
    setInstallDialogOpen(false);
    setSelectedAgents(new Set());
    setSelectedSkillsForInstall([]);
    setSelectedPoolSkills(new Set());
    setInstalling(false);
    onRefresh();
  };

  const toggleSelectAgent = (agentId: string) => {
    setSelectedAgents(prev => {
      const next = new Set(prev);
      next.has(agentId) ? next.delete(agentId) : next.add(agentId);
      return next;
    });
  };

  const openInstallDialog = (skillOrSkills: ListedSkill | ListedSkill[]) => {
    const skills = Array.isArray(skillOrSkills) ? skillOrSkills : [skillOrSkills];
    setSelectedSkillsForInstall(skills);
    // Auto-select detected agents
    const detectedIds = agents.filter(a => a.detected).map(a => a.id);
    setSelectedAgents(new Set(detectedIds));
    setInstallModeTab("agent");
    setProjectInstallPath("");
    setInstallDialogOpen(true);
  };

  const toggleSelectPoolSkill = (skillName: string) => {
    setSelectedPoolSkills(prev => {
      const next = new Set(prev);
      next.has(skillName) ? next.delete(skillName) : next.add(skillName);
      return next;
    });
  };

  const toggleSelectAllPool = () => {
    const allNames = filteredPoolSkills.map(s => s.name);
    if (selectedPoolSkills.size === allNames.length && allNames.length > 0) {
      setSelectedPoolSkills(new Set());
    } else {
      setSelectedPoolSkills(new Set(allNames));
    }
  };

  const handleBatchInstall = () => {
    const selected = poolSkillsFromIndex.filter(s => selectedPoolSkills.has(s.name));
    if (selected.length === 0) return;
    openInstallDialog(selected);
  };

  const handleInstallToProject = async () => {
    if (selectedSkillsForInstall.length === 0 || !projectInstallPath.trim() || selectedAgents.size === 0) return;
    setInstalling(true);
    const results: InstallResultItem[] = [];
    for (const skill of selectedSkillsForInstall) {
      for (const agentId of selectedAgents) {
        const agent = agents.find(a => a.id === agentId);
        try {
          await installToProjectForAgent(skill.storePath || skill.paths[0], projectInstallPath.trim(), agentId, true);
          results.push({ skillName: skill.name, agentName: agent?.name || agentId, success: true });
        } catch (e: any) {
          const errMsg = typeof e === "string" ? e : (e?.message || e?.toString?.() || String(e));
          results.push({ skillName: skill.name, agentName: agent?.name || agentId, success: false, error: errMsg });
        }
      }
    }
    const totalSuccess = results.filter(r => r.success).length;
    const totalFail = results.filter(r => !r.success).length;

    if (totalFail === 0) {
      toast.success(`成功安装 ${selectedSkillsForInstall.length} 个技能到 ${selectedAgents.size} 个智能体的项目目录`);
    } else {
      toast.error(`安装完成：${totalSuccess} 成功，${totalFail} 失败`);
      setInstallResults(results);
      setInstallResultOpen(true);
    }
    setInstallDialogOpen(false);
    setSelectedSkillsForInstall([]);
    setSelectedPoolSkills(new Set());
    setSelectedAgents(new Set());
    setProjectInstallPath("");
    setInstalling(false);
    onRefresh();
  };

  const handleOpenFolder = async (path: string) => {
    setOpeningFolder(path);
    try {
      await openDirectory(path);
    } catch { /* ignore */ }
    finally { setOpeningFolder(null); }
  };

  const handleScan = async () => {
    setScanning(true);
    setScanError(null);
    try {
      const res = await scanLocal(projectPath.trim() || undefined);
      setScanResults(res);
      if (res.length === 0) {
        setScanError("未发现技能。请确认路径下包含 SKILL.md 的技能目录。");
      }
      onRefresh();
    } catch (e: any) {
      setScanError(e.message || "扫描失败。");
    } finally {
      setScanning(false);
    }
  };

  const handleBrowseProjectPath = async () => {
    try {
      const dir = await openDirectoryDialog("选择扫描目录");
      if (dir) setProjectPath(dir);
    } catch { /* canceled */ }
  };

  const handleImport = async (skill: MergedScanSkill) => {
    setImporting(true);
    let s = 0, f = 0;
    for (const path of skill.paths) {
      try { await importToPool(path); s++; } catch { f++; }
    }
    toast[f === 0 ? "success" : "error"](`复制到池：${s} 个成功${f > 0 ? `，${f} 个失败` : ""}`);
    onRefresh();
    setImporting(false);
  };

  const handleArchive = async (skill: MergedScanSkill) => {
    setImporting(true);
    let s = 0, f = 0;
    for (const path of skill.paths) {
      try { await archiveToPool(path); s++; } catch { f++; }
    }
    toast[f === 0 ? "success" : "error"](`归档到池：${s} 个成功${f > 0 ? `，${f} 个失败` : ""}`);
    onRefresh();
    setImporting(false);
  };

  const handleBatchCopyToPool = async () => {
    if (selectedScanSkills.size === 0) return;
    setImporting(true);
    let s = 0, f = 0;
    for (const name of selectedScanSkills) {
      const skill = mergedScanResults.find(r => r.name === name);
      if (!skill) { f++; continue; }
      for (const path of skill.paths) {
        try { await importToPool(path); s++; } catch { f++; }
      }
    }
    toast[f === 0 ? "success" : "error"](`批量复制完成：${s} 个成功${f > 0 ? `，${f} 个失败` : ""}`);
    setSelectedScanSkills(new Set());
    onRefresh();
    setImporting(false);
  };

  const handleBatchArchiveToPool = async () => {
    if (selectedScanSkills.size === 0) return;
    setImporting(true);
    let s = 0, f = 0;
    for (const name of selectedScanSkills) {
      const skill = mergedScanResults.find(r => r.name === name);
      if (!skill) { f++; continue; }
      for (const path of skill.paths) {
        try { await archiveToPool(path); s++; } catch { f++; }
      }
    }
    toast[f === 0 ? "success" : "error"](`批量归档完成：${s} 个成功${f > 0 ? `，${f} 个失败` : ""}`);
    setSelectedScanSkills(new Set());
    onRefresh();
    setImporting(false);
  };

  const handleBatchDeleteScan = async () => {
    if (selectedScanSkills.size === 0) return;
    setDeleting(true);
    let s = 0, f = 0;
    for (const name of selectedScanSkills) {
      const skill = mergedScanResults.find(r => r.name === name);
      if (!skill) { f++; continue; }
      for (const path of skill.paths) {
        try { await deleteSkillFromProject(path); s++; } catch { f++; }
      }
    }
    toast[f === 0 ? "success" : "error"](`批量删除完成：${s} 个成功${f > 0 ? `，${f} 个失败` : ""}`);
    setSelectedScanSkills(new Set());
    setDeleting(false);
    onRefresh();
  };

  const handleImportToListed = async (skill: ListedSkill) => {
    // Prefer storePath (repo path) over agent symlink path for archiving
    const sourcePath = skill.storePath || (skill.paths.length > 0 ? skill.paths[0] : "");
    if (!sourcePath) return;
    setImporting(true);
    try {
      await archiveToPool(sourcePath);
      toast.success(`"${skill.name}" 已归档到技能池`);
      onRefresh();
    } catch (e: any) {
      const errMsg = typeof e === "string" ? e : (e?.message || e?.toString?.() || String(e));
      toast.error(`归档失败: ${errMsg}`);
    } finally {
      setImporting(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    try {
      if (deleteTarget.level === "pool") {
        await deleteSkillFromPool(deleteTarget.skill.name);
        toast.success(`"${deleteTarget.skill.name}" 已从技能池及所有智能体中彻底删除`);
      } else if (deleteTarget.level === "agent") {
        const pathsToDelete = deleteTarget.paths || [deleteTarget.path];
        let successCount = 0;
        for (const p of pathsToDelete) {
          try { await deleteSkillFromAgent(p); successCount++; } catch { /* continue */ }
        }
        toast.success(`已从 ${successCount} 个智能体全局目录删除 "${deleteTarget.skill.name}"（技能池不受影响）`);
      } else if (deleteTarget.level === "project") {
        await deleteSkillFromProject(deleteTarget.path);
        toast.success(`已从项目删除 "${deleteTarget.skill.name}"`);
      }
      setDeleteTarget(null);
      onRefresh();
    } catch (e: any) {
      const errMsg = typeof e === "string" ? e : (e?.message || e?.toString?.() || String(e));
      toast.error(`删除失败: ${errMsg}`);
    } finally {
      setDeleting(false);
    }
  };

  const toggleSelectGlobal = (name: string) => {
    setSelectedGlobalSkills(prev => {
      const next = new Set(prev);
      next.has(name) ? next.delete(name) : next.add(name);
      return next;
    });
  };

  const toggleSelectAllGlobal = () => {
    const allNames = filteredGlobalSkills.map(s => s.name);
    if (selectedGlobalSkills.size === allNames.length && allNames.length > 0) {
      setSelectedGlobalSkills(new Set());
    } else {
      setSelectedGlobalSkills(new Set(allNames));
    }
  };

  const handleBatchImportToPool = async () => {
    if (selectedGlobalSkills.size === 0) return;
    setImporting(true);
    let s = 0, f = 0;
    for (const name of selectedGlobalSkills) {
      const skill = skills.find(sk => sk.name === name);
      if (!skill) { f++; continue; }
      const sourcePath = skill.storePath || (skill.paths.length > 0 ? skill.paths[0] : "");
      if (!sourcePath) { f++; continue; }
      try { await archiveToPool(sourcePath); s++; } catch { f++; }
    }
    toast[f === 0 ? "success" : "error"](`批量入池完成：${s} 个成功${f > 0 ? `，${f} 个失败` : ""}`);
    setSelectedGlobalSkills(new Set());
    onRefresh();
    setImporting(false);
  };

  // Pool skills from index
  const poolSkillsFromIndex = skills.filter(s => s.inPool);

  // Filtered pool skills by keyword (match name, agent names, paths)
  const filteredPoolSkills = poolKeyword.trim()
    ? poolSkillsFromIndex.filter(s => {
        const kw = poolKeyword.toLowerCase();
        return (
          s.name.toLowerCase().includes(kw) ||
          s.agentNames.some(n => n.toLowerCase().includes(kw)) ||
          s.paths.some(p => p.toLowerCase().includes(kw))
        );
      })
    : poolSkillsFromIndex;

  // Filtered global skills (only skills with at least one agent, match name/agent/path)
  const globalSkillsWithAgent = skills.filter(s => s.agentNames.length > 0);
  const filteredGlobalSkills = globalKeyword.trim()
    ? globalSkillsWithAgent.filter(s => {
        const kw = globalKeyword.toLowerCase();
        return (
          s.name.toLowerCase().includes(kw) ||
          s.agentNames.some(n => n.toLowerCase().includes(kw)) ||
          s.paths.some(p => p.toLowerCase().includes(kw))
        );
      })
    : globalSkillsWithAgent;

  // Scan results: merge by skill name, aggregate agent info
  interface MergedScanSkill {
    name: string;
    agentNames: string[];
    agentIds: string[];
    paths: string[];
    inPool: boolean;
  }
  const mergedScanResults: MergedScanSkill[] = (() => {
    const map = new Map<string, MergedScanSkill>();
    for (const s of scanResults) {
      const existing = map.get(s.name);
      if (existing) {
        if (s.agentName && !existing.agentNames.includes(s.agentName)) {
          existing.agentNames.push(s.agentName);
          existing.agentIds.push(s.agentId || "");
        }
        if (!existing.paths.includes(s.path)) {
          existing.paths.push(s.path);
        }
        if (s.alreadyInPool) existing.inPool = true;
      } else {
        map.set(s.name, {
          name: s.name,
          agentNames: s.agentName ? [s.agentName] : [],
          agentIds: s.agentId ? [s.agentId] : [],
          paths: [s.path],
          inPool: s.alreadyInPool,
        });
      }
    }
    return Array.from(map.values()).sort((a, b) => a.name.localeCompare(b.name));
  })();

  const toggleSelectScanSkill = (name: string) => {
    setSelectedScanSkills(prev => {
      const next = new Set(prev);
      next.has(name) ? next.delete(name) : next.add(name);
      return next;
    });
  };

  const toggleSelectAllScan = () => {
    const allNames = mergedScanResults.map(s => s.name);
    if (selectedScanSkills.size === allNames.length && allNames.length > 0) {
      setSelectedScanSkills(new Set());
    } else {
      setSelectedScanSkills(new Set(allNames));
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">技能池</h2>
        <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading}>
          {loading ? <><Loader2 className="h-4 w-4 animate-spin" /> 刷新中...</> : <><RefreshCw className="h-4 w-4" /> 刷新</>}
        </Button>
      </div>

      <Tabs defaultValue="pool">
        <TabsList>
          <TabsTrigger value="pool">本机技能池 ({poolKeyword.trim() ? `${filteredPoolSkills.length}/${poolSkillsFromIndex.length}` : poolSkillsFromIndex.length})</TabsTrigger>
          <TabsTrigger value="global">智能体全局技能 ({globalSkillsWithAgent.length})</TabsTrigger>
          <TabsTrigger value="scan">本机项目技能</TabsTrigger>
        </TabsList>

        {/* Tab 1: Pool Skills */}
        <TabsContent value="pool">
          <Card>
            <CardHeader>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center gap-2 text-sm">
                    <CheckCircle2 className="h-4 w-4 text-green-500" />
                    技能池
                    {poolPath && !loadingPool && <span className="text-muted-foreground font-normal">({poolPath})</span>}
                  </CardTitle>
                  <div className="flex items-center gap-2">
                    {selectedPoolSkills.size > 0 && (
                      <Button variant="default" size="sm" className="h-7 text-xs" onClick={handleBatchInstall} disabled={installing}>
                        {installing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Plus className="h-3 w-3" />}
                        批量安装 ({selectedPoolSkills.size})
                      </Button>
                    )}
                    <Button variant="ghost" size="sm" className="h-7 text-xs gap-1" onClick={toggleSelectAllPool}>
                      {filteredPoolSkills.length > 0 && selectedPoolSkills.size === filteredPoolSkills.length ? <CheckCircle2 className="h-3.5 w-3.5" /> : <Circle className="h-3.5 w-3.5 text-muted-foreground" />}
                      {filteredPoolSkills.length > 0 && selectedPoolSkills.size === filteredPoolSkills.length ? "取消全选" : "全选"}
                    </Button>
                  </div>
                </div>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    className="pl-9"
                    placeholder="筛选技能名称、智能体或路径..."
                    value={poolKeyword}
                    onChange={(e) => { setPoolKeyword(e.target.value); setSelectedPoolSkills(new Set()); }}
                  />
                </div>
                {poolKeyword.trim() && (
                  <p className="text-xs text-muted-foreground">
                    筛选 "{poolKeyword}"：{filteredPoolSkills.length} / {poolSkillsFromIndex.length} 个技能
                  </p>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {filteredPoolSkills.length === 0 ? (
                <p className="text-muted-foreground text-sm py-4 text-center">
                  {poolKeyword.trim() ? "无匹配技能" : "技能池为空"}
                </p>
              ) : (
                <div className="border rounded-lg">
                  <div className="max-h-[60vh] overflow-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="bg-muted/50 border-b">
                        <th className="text-left font-medium px-4 py-2 w-8">
                          <button className="p-0.5 rounded hover:bg-muted" onClick={(e) => { e.stopPropagation(); toggleSelectAllPool(); }} title="全选/取消全选">
                            {filteredPoolSkills.length > 0 && selectedPoolSkills.size === filteredPoolSkills.length ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <Circle className="h-4 w-4 text-muted-foreground" />}
                          </button>
                        </th>
                        <th className="text-left font-medium px-4 py-2">技能名称</th>
                        <th className="text-left font-medium px-4 py-2">池路径</th>
                        <th className="text-right font-medium px-4 py-2">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredPoolSkills.map((s) => {
                        const isSelected = selectedPoolSkills.has(s.name);
                        const poolSkillPath = poolPath ? `${poolPath}/${s.name}` : "";
                        return (
                        <tr key={s.name} className={`border-b last:border-0 hover:bg-muted/30`} onClick={() => onSelect(s)}>
                          <td className="px-4 py-2" onClick={(e) => e.stopPropagation()}>
                            <button
                              className="p-0.5 rounded hover:bg-muted"
                              onClick={() => toggleSelectPoolSkill(s.name)}
                            >
                              {isSelected ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <Circle className="h-4 w-4 text-muted-foreground" />}
                            </button>
                          </td>
                          <td className="px-4 py-2 font-medium">{s.name}</td>
                          <td className="px-4 py-2">
                            {poolSkillPath ? (
                              <div className="flex items-center gap-1">
                                <span className="text-xs text-muted-foreground font-mono truncate max-w-[200px]" title={poolSkillPath}>{poolSkillPath}</span>
                                <Button variant="ghost" size="icon" className="h-5 w-5 shrink-0" onClick={(e) => { e.stopPropagation(); handleOpenFolder(poolSkillPath); }} disabled={openingFolder === poolSkillPath}>
                                  {openingFolder === poolSkillPath ? <Loader2 className="h-3 w-3 animate-spin" /> : <ExternalLink className="h-3 w-3" />}
                                </Button>
                              </div>
                            ) : <span className="text-xs text-muted-foreground">-</span>}
                          </td>
                          <td className="px-4 py-2 text-right">
                            <div className="flex justify-end gap-1">
                              <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); onSelect(s); }}>详情</Button>
                              <Button variant="outline" size="sm" className="h-7 text-xs" onClick={(e) => { e.stopPropagation(); openInstallDialog(s); }} disabled={installing}>
                                {installing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
                                安装
                              </Button>
                              <Button variant="ghost" size="sm" className="h-7 text-xs text-destructive hover:text-destructive" onClick={(e) => { e.stopPropagation(); setDeleteTarget({ level: "pool", skill: s, path: poolSkillPath }); }} disabled={deleting}>
                                <Trash2 className="h-3.5 w-3.5" />
                              </Button>
                            </div>
                          </td>
                        </tr>
                      )})}
                    </tbody>
                  </table>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Tab 2: Global Agent Skills */}
        <TabsContent value="global">
          <Card>
            <CardHeader>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center gap-2 text-sm">
                    <Puzzle className="h-4 w-4 text-primary" />
                    智能体全局技能
                  </CardTitle>
                  <div className="flex items-center gap-2">
                    {selectedGlobalSkills.size > 0 && (
                      <Button variant="default" size="sm" className="h-7 text-xs" onClick={handleBatchImportToPool} disabled={importing}>
                        {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Download className="h-3 w-3" />}
                        批量入池 ({selectedGlobalSkills.size})
                      </Button>
                    )}
                    <Button variant="ghost" size="sm" className="h-7 text-xs gap-1" onClick={toggleSelectAllGlobal}>
                      {filteredGlobalSkills.length > 0 && selectedGlobalSkills.size === filteredGlobalSkills.length ? <CheckCircle2 className="h-3.5 w-3.5" /> : <Circle className="h-3.5 w-3.5 text-muted-foreground" />}
                      {filteredGlobalSkills.length > 0 && selectedGlobalSkills.size === filteredGlobalSkills.length ? "取消全选" : "全选"}
                    </Button>
                  </div>
                </div>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    className="pl-9"
                    placeholder="筛选技能名称、智能体或路径..."
                    value={globalKeyword}
                    onChange={(e) => { setGlobalKeyword(e.target.value); setSelectedGlobalSkills(new Set()); }}
                  />
                </div>
                {globalKeyword.trim() && (
                  <p className="text-xs text-muted-foreground">
                    筛选 "{globalKeyword}"：{filteredGlobalSkills.length} / {globalSkillsWithAgent.length} 个技能
                  </p>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {filteredGlobalSkills.length === 0 ? (
                <p className="text-muted-foreground text-sm py-4 text-center">
                  {globalKeyword.trim() ? "无匹配技能" : "暂无技能数据"}
                </p>
              ) : (
                <div className="border rounded-lg">
                  <div className="max-h-[60vh] overflow-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="bg-muted/50 border-b">
                        <th className="text-left font-medium px-4 py-2 w-8">
                          <button className="p-0.5 rounded hover:bg-muted" onClick={(e) => { e.stopPropagation(); toggleSelectAllGlobal(); }} title="全选/取消全选">
                            {filteredGlobalSkills.length > 0 && selectedGlobalSkills.size === filteredGlobalSkills.length ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <Circle className="h-4 w-4 text-muted-foreground" />}
                          </button>
                        </th>
                        <th className="text-left font-medium px-4 py-2">技能名称</th>
                        <th className="text-left font-medium px-4 py-2">智能体</th>
                        <th className="text-right font-medium px-4 py-2">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredGlobalSkills.map((s) => {
                        return (
                        <tr key={s.name} className={`border-b last:border-0 hover:bg-muted/30`}>
                          <td className="px-4 py-2">
                            <button
                              className="p-0.5 rounded hover:bg-muted"
                              onClick={() => toggleSelectGlobal(s.name)}
                            >
                              {selectedGlobalSkills.has(s.name) ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <Circle className="h-4 w-4 text-muted-foreground" />}
                            </button>
                          </td>
                          <td className="px-4 py-2 font-medium">
                            <div className="flex items-center gap-1.5">
                              {s.name}
                              {s.inPool && <Badge variant="default" className="text-[10px]">池</Badge>}
                            </div>
                          </td>
                          <td className="px-4 py-2">
                            <div className="flex flex-wrap gap-1">
                              {s.agentNames.map((name, i) => (
                                <Badge key={i} variant="secondary" className="text-[10px]">{name}</Badge>
                              ))}
                            </div>
                          </td>
                          <td className="px-4 py-2 text-right">
                            <div className="flex justify-end gap-1">
                              {s.inPool ? (
                                <Button variant="outline" size="sm" className="h-7 text-xs" onClick={(e) => { e.stopPropagation(); openInstallDialog(s); }} disabled={installing}>
                                  {installing ? <><Loader2 className="h-3 w-3 animate-spin" /> 安装中...</> : <><Plus className="h-3.5 w-3.5" /> 安装</>}
                                </Button>
                              ) : (
                                <Button variant="outline" size="sm" className="h-7 text-xs" onClick={(e) => { e.stopPropagation(); handleImportToListed(s); }} disabled={importing || (!s.storePath && s.paths.length === 0)}>
                                  {importing ? <><Loader2 className="h-3 w-3 animate-spin" /> 归档中...</> : <><Download className="h-3.5 w-3.5" /> 归入技能池</>}
                                </Button>
                              )}
                              <Button variant="ghost" size="sm" className="h-7 text-xs text-destructive hover:text-destructive" onClick={(e) => { e.stopPropagation(); setDeleteTarget({ level: "agent", skill: s, path: s.paths[0], paths: s.paths }); }} disabled={deleting}>
                                <Trash2 className="h-3.5 w-3.5" />
                              </Button>
                            </div>
                          </td>
                        </tr>
                      )})}
                    </tbody>
                  </table>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Tab 3: Project Skills (Scan Results) */}
        <TabsContent value="scan">
          {/* Scan Input */}
          <Card>
            <CardHeader>
              <CardTitle>本机项目技能</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-2">
                <Input
                  className="flex-1"
                  placeholder="选择要扫描的项目目录（留空则扫描所有智能体全局目录）"
                  value={projectPath}
                  onChange={(e) => setProjectPath(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleScan()}
                />
                <Button variant="outline" size="sm" onClick={handleBrowseProjectPath}>
                  <FolderOpen className="h-3.5 w-3.5" /> 浏览
                </Button>
                <Button onClick={handleScan} disabled={scanning}>
                  {scanning ? <><Loader2 className="h-4 w-4 animate-spin" /> 扫描中...</> : <><Search className="h-4 w-4" /> 扫描</>}
                </Button>
              </div>
            </CardContent>
          </Card>

          {scanError && (
            <Card>
              <CardContent className="py-4">
                <p className="text-sm text-destructive">{scanError}</p>
              </CardContent>
            </Card>
          )}

          {/* Merged results table */}
          {mergedScanResults.length > 0 && (
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center gap-2 text-sm">
                    <Puzzle className="h-4 w-4 text-primary" />
                    项目技能 ({mergedScanResults.length})
                  </CardTitle>
                  <div className="flex items-center gap-2">
                    {selectedScanSkills.size > 0 && (
                      <>
                        <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleBatchCopyToPool} disabled={importing}>
                          {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Download className="h-3 w-3" />}
                          复制到池 ({selectedScanSkills.size})
                        </Button>
                        <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleBatchArchiveToPool} disabled={importing}>
                          {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Archive className="h-3 w-3" />}
                          归档到池 ({selectedScanSkills.size})
                        </Button>
                        <Button variant="ghost" size="sm" className="h-7 text-xs text-destructive hover:text-destructive" onClick={handleBatchDeleteScan} disabled={deleting}>
                          {deleting ? <Loader2 className="h-3 w-3 animate-spin" /> : <Trash2 className="h-3 w-3" />}
                          删除 ({selectedScanSkills.size})
                        </Button>
                      </>
                    )}
                    <TooltipProvider delayDuration={0}>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <button className="p-1 rounded hover:bg-muted" type="button">
                            <HelpCircle className="h-3.5 w-3.5 text-muted-foreground" />
                          </button>
                        </TooltipTrigger>
                        <TooltipContent side="left" className="max-w-xs">
                          <div className="space-y-1 text-xs">
                            <p><strong>复制到池</strong>：将技能复制到技能池，原目录保留不动</p>
                            <p><strong>归档到池</strong>：将技能移入技能池，原目录删除（相当于搬家）</p>
                            <p><strong>删除</strong>：仅删除该项目中的技能，不影响技能池</p>
                          </div>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                    <Button variant="ghost" size="sm" className="h-7 text-xs gap-1" onClick={toggleSelectAllScan}>
                      {mergedScanResults.length > 0 && selectedScanSkills.size === mergedScanResults.length ? <CheckCircle2 className="h-3.5 w-3.5" /> : <Circle className="h-3.5 w-3.5 text-muted-foreground" />}
                      {mergedScanResults.length > 0 && selectedScanSkills.size === mergedScanResults.length ? "取消全选" : "全选"}
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="border rounded-lg">
                  <div className="max-h-[60vh] overflow-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="bg-muted/50 border-b">
                        <th className="text-left font-medium px-4 py-2 w-8">
                          <button className="p-0.5 rounded hover:bg-muted" onClick={(e) => { e.stopPropagation(); toggleSelectAllScan(); }} title="全选/取消全选">
                            {mergedScanResults.length > 0 && selectedScanSkills.size === mergedScanResults.length ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <Circle className="h-4 w-4 text-muted-foreground" />}
                          </button>
                        </th>
                        <th className="text-left font-medium px-4 py-2">技能名称</th>
                        <th className="text-left font-medium px-4 py-2">智能体</th>
                        <th className="text-right font-medium px-4 py-2">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {mergedScanResults.map((s) => {
                        const isSelected = selectedScanSkills.has(s.name);
                        return (
                        <tr key={s.name} className="border-b last:border-0 hover:bg-muted/30">
                          <td className="px-4 py-2">
                            <button
                              className="p-0.5 rounded hover:bg-muted"
                              onClick={() => toggleSelectScanSkill(s.name)}
                            >
                              {isSelected ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <Circle className="h-4 w-4 text-muted-foreground" />}
                            </button>
                          </td>
                          <td className="px-4 py-2 font-medium">
                            <div className="flex items-center gap-1.5">
                              {s.name}
                              {s.inPool && <Badge variant="default" className="text-[10px]">池</Badge>}
                            </div>
                          </td>
                          <td className="px-4 py-2">
                            <div className="flex flex-wrap gap-1">
                              {s.agentNames.map((name, i) => (
                                <Badge key={i} variant="secondary" className="text-[10px]">{name}</Badge>
                              ))}
                              {s.agentNames.length === 0 && <span className="text-xs text-muted-foreground">-</span>}
                            </div>
                          </td>
                          <td className="px-4 py-2 text-right">
                            <div className="flex justify-end gap-1">
                              {!s.inPool && (
                                <>
                                  <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => handleImport(s)} disabled={importing}>
                                    {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Download className="h-3 w-3" />}
                                    复制到池
                                  </Button>
                                  <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => handleArchive(s)} disabled={importing}>
                                    {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Archive className="h-3 w-3" />}
                                    归档到池
                                  </Button>
                                </>
                              )}
                              <Button variant="ghost" size="sm" className="h-7 text-xs text-destructive hover:text-destructive" onClick={() => setDeleteTarget({ level: "project", skill: { name: s.name } as any, path: s.paths[0], paths: s.paths })} disabled={deleting}>
                                <Trash2 className="h-3.5 w-3.5" />
                              </Button>
                            </div>
                          </td>
                        </tr>
                      )})}
                    </tbody>
                  </table>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {scanResults.length === 0 && !scanError && !scanning && projectPath && (
            <Card>
              <CardContent className="py-4">
                <p className="text-sm text-muted-foreground text-center">点击"扫描"查找项目中的技能</p>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>

      {/* Install Dialog */}
      <Dialog open={installDialogOpen} onOpenChange={setInstallDialogOpen}>
        <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle>安装技能</DialogTitle>
          </DialogHeader>
          <Tabs value={installModeTab} onValueChange={setInstallModeTab} className="flex-1 flex flex-col overflow-hidden">
            <TabsList className="grid grid-cols-2">
              <TabsTrigger value="agent">智能体全局安装</TabsTrigger>
              <TabsTrigger value="project">项目级别安装</TabsTrigger>
            </TabsList>

            {/* Agent global install tab */}
            <TabsContent value="agent" className="flex-1 overflow-y-auto space-y-4 py-4">
              <div>
                <Label className="text-sm font-medium">
                  安装技能 ({selectedSkillsForInstall.length})
                </Label>
                <div className="mt-1 max-h-32 overflow-y-auto border rounded-md p-2 space-y-1">
                  {selectedSkillsForInstall.map((skill, i) => (
                    <div key={i} className="flex items-center gap-2 text-sm px-2 py-1 rounded bg-muted/50">
                      <Puzzle className="h-3.5 w-3.5 text-primary shrink-0" />
                      <span className="font-medium">{skill.name}</span>
                    </div>
                  ))}
                </div>
              </div>
              <div>
                <Label className="text-sm font-medium">选择智能体</Label>
                <div className="mt-2 max-h-48 overflow-y-auto space-y-1 border rounded-md p-3">
                  {agents.length === 0 ? (
                    <p className="text-xs text-muted-foreground text-center py-3">未检测到智能体</p>
                  ) : (
                    agents.map((agent) => (
                      <label key={agent.id} className="flex items-center gap-2 p-2 rounded hover:bg-muted cursor-pointer">
                        <input
                          type="checkbox"
                          checked={selectedAgents.has(agent.id)}
                          onChange={() => toggleSelectAgent(agent.id)}
                          className="h-4 w-4 rounded border-gray-300"
                        />
                        <div className="flex-1 min-w-0">
                          <div className="text-sm font-medium flex items-center gap-2">
                            {agent.name}
                            {agent.detected && <Badge variant="default" className="text-[10px]">已检测</Badge>}
                          </div>
                          {agent.skillsDir && (
                            <p className="text-xs text-muted-foreground font-mono truncate">{agent.skillsDir}</p>
                          )}
                        </div>
                      </label>
                    ))
                  )}
                </div>
              </div>
              {selectedAgents.size > 0 && (
                <p className="text-xs text-muted-foreground">
                  将安装 {selectedSkillsForInstall.length} 个技能到 {selectedAgents.size} 个智能体（目录存在则覆盖）
                </p>
              )}
              <div className="flex justify-end gap-2 pt-2 border-t">
                <Button variant="outline" onClick={() => setInstallDialogOpen(false)} disabled={installing}>取消</Button>
                <Button onClick={handleInstall} disabled={installing || selectedAgents.size === 0}>
                  {installing ? <><Loader2 className="h-4 w-4 animate-spin" /> 安装中...</> : <><Plus className="h-4 w-4" /> 确认安装</>}
                </Button>
              </div>
            </TabsContent>

            {/* Project-level install tab */}
            <TabsContent value="project" className="flex-1 overflow-y-auto space-y-4 py-4">
              <div>
                <Label className="text-sm font-medium">
                  安装技能 ({selectedSkillsForInstall.length})
                </Label>
                <div className="mt-1 max-h-32 overflow-y-auto border rounded-md p-2 space-y-1">
                  {selectedSkillsForInstall.map((skill, i) => (
                    <div key={i} className="flex items-center gap-2 text-sm px-2 py-1 rounded bg-muted/50">
                      <Puzzle className="h-3.5 w-3.5 text-primary shrink-0" />
                      <span className="font-medium">{skill.name}</span>
                    </div>
                  ))}
                </div>
              </div>
              <div>
                <Label className="text-sm font-medium">选择项目目录</Label>
                <div className="flex gap-2 mt-2">
                  <Input
                    className="flex-1"
                    placeholder="选择项目目录..."
                    value={projectInstallPath}
                    onChange={(e) => setProjectInstallPath(e.target.value)}
                  />
                  <Button variant="outline" size="sm" onClick={async () => {
                    const dir = await openDirectoryDialog("选择项目目录");
                    if (dir) setProjectInstallPath(dir);
                  }}>
                    <FolderOpen className="h-3.5 w-3.5" /> 浏览
                  </Button>
                </div>
              </div>
              <div>
                <Label className="text-sm font-medium">选择智能体（每个智能体安装到各自的项目级目录）</Label>
                <div className="mt-2 max-h-48 overflow-y-auto space-y-1 border rounded-md p-3">
                  {agents.length === 0 ? (
                    <p className="text-xs text-muted-foreground text-center py-3">未检测到智能体</p>
                  ) : (
                    agents.map((agent) => (
                      <label key={agent.id} className="flex items-center gap-2 p-2 rounded hover:bg-muted cursor-pointer">
                        <input
                          type="checkbox"
                          checked={selectedAgents.has(agent.id)}
                          onChange={() => toggleSelectAgent(agent.id)}
                          className="h-4 w-4 rounded border-gray-300"
                        />
                        <div className="flex-1 min-w-0">
                          <div className="text-sm font-medium flex items-center gap-2">
                            {agent.name}
                            {agent.detected && <Badge variant="default" className="text-[10px]">已检测</Badge>}
                          </div>
                          {agent.projectSkillsSubdir && (
                            <p className="text-xs text-muted-foreground font-mono">
                              {projectInstallPath.trim()
                                ? `${projectInstallPath}/${agent.projectSkillsSubdir}`
                                : `<项目>/${agent.projectSkillsSubdir}`}
                            </p>
                          )}
                        </div>
                      </label>
                    ))
                  )}
                </div>
              </div>
              {selectedAgents.size > 0 && projectInstallPath.trim() && (
                <p className="text-xs text-muted-foreground">
                  将安装 {selectedSkillsForInstall.length} 个技能到 {selectedAgents.size} 个智能体的项目级目录（目录存在则覆盖）
                </p>
              )}
              <div className="flex justify-end gap-2 pt-2 border-t">
                <Button variant="outline" onClick={() => setInstallDialogOpen(false)} disabled={installing}>取消</Button>
                <Button onClick={handleInstallToProject} disabled={installing || !projectInstallPath.trim() || selectedAgents.size === 0}>
                  {installing ? <><Loader2 className="h-4 w-4 animate-spin" /> 安装中...</> : <><Plus className="h-4 w-4" /> 确认安装</>}
                </Button>
              </div>
            </TabsContent>
          </Tabs>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation dialog */}
      <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => { if (!open) setDeleteTarget(null); }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除技能</AlertDialogTitle>
            <AlertDialogDescription asChild>
              <div className="space-y-2">
                {deleteTarget?.level === "pool" && (
                  <>
                    <p>确定要从技能池中彻底删除 <strong className="text-foreground">"{deleteTarget?.skill?.name}"</strong> 吗？</p>
                    <p className="text-destructive font-medium">这是机器级操作，将同时删除：</p>
                    <ul className="list-disc list-inside text-sm space-y-1">
                      <li>技能池中的文件</li>
                      {(deleteTarget?.skill as ListedSkill)?.agentNames?.length > 0 && (
                        <li>以下智能体全局目录中的引用：{(deleteTarget.skill as ListedSkill).agentNames.join("、")}</li>
                      )}
                      <li>所有项目中该技能的引用</li>
                    </ul>
                  </>
                )}
                {deleteTarget?.level === "agent" && (
                  <>
                    <p>确定要从智能体全局目录中删除 <strong className="text-foreground">"{deleteTarget?.skill?.name}"</strong> 吗？</p>
                    <p>此操作将删除以下智能体全局目录中的引用：</p>
                    <ul className="list-disc list-inside text-sm space-y-1">
                      {(deleteTarget?.skill as ListedSkill)?.agentNames?.map((name, i) => (
                        <li key={i}>{name}</li>
                      ))}
                    </ul>
                    <p className="text-green-600 dark:text-green-400 font-medium">技能池中的技能文件不受影响。</p>
                  </>
                )}
                {deleteTarget?.level === "project" && (
                  <>
                    <p>确定要从项目中删除 <strong className="text-foreground">"{deleteTarget?.skill?.name}"</strong> 吗？</p>
                    <p>此操作仅删除该项目中的技能，不影响技能池和其他智能体。</p>
                  </>
                )}
                <p className="text-destructive font-medium">此操作不可撤销。</p>
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleting}>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={deleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleting ? <><Loader2 className="h-4 w-4 animate-spin" /> 删除中...</> : <><Trash2 className="h-4 w-4" /> 确认删除</>}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Install Result Dialog */}
      <Dialog open={installResultOpen} onOpenChange={setInstallResultOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>安装结果详情</DialogTitle>
          </DialogHeader>
          <div className="max-h-[60vh] overflow-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="text-left font-medium px-3 py-1.5">技能</th>
                  <th className="text-left font-medium px-3 py-1.5">智能体</th>
                  <th className="text-left font-medium px-3 py-1.5">结果</th>
                </tr>
              </thead>
              <tbody>
                {installResults.map((r, i) => (
                  <tr key={i} className="border-b last:border-0">
                    <td className="px-3 py-1.5">{r.skillName}</td>
                    <td className="px-3 py-1.5">{r.agentName}</td>
                    <td className="px-3 py-1.5">
                      {r.success ? (
                        <Badge variant="default" className="text-[10px]">成功</Badge>
                      ) : (
                        <div className="space-y-0.5">
                          <Badge variant="destructive" className="text-[10px]">失败</Badge>
                          <p className="text-xs text-destructive mt-0.5">{r.error}</p>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="flex justify-end gap-2 mt-2">
            <Button variant="outline" size="sm" onClick={() => setInstallResultOpen(false)}>关闭</Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
