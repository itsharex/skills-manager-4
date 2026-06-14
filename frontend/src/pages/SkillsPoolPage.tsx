import { useState, useEffect } from "react";
import { RefreshCw, Search, Loader2, Puzzle, CheckCircle2, FolderOpen, Download, Archive, ExternalLink, Plus } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "../components/ui/tabs";
import { scanLocal, getConfig, importToPool, archiveToPool, openDirectoryDialog, installToAgent, listAgents, openDirectory, installToProject } from "../bridge";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "../components/ui/dialog";
import { Label } from "../components/ui/label";
import type { ListedSkill, DiscoveredSkill, AgentInfo } from "../types";

interface Props {
  skills: ListedSkill[];
  onSelect: (skill: ListedSkill) => void;
  onRefresh: () => void;
}

export default function SkillsPoolPage({ skills, onSelect, onRefresh }: Props) {
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
  const [actionMsg, setActionMsg] = useState<{ success: boolean; msg: string } | null>(null);

  // Multi-select state for scan results
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set());

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
    let totalSuccess = 0, totalFail = 0;
    for (const skill of selectedSkillsForInstall) {
      for (const agentId of selectedAgents) {
        const agent = agents.find(a => a.id === agentId);
        if (!agent || !agent.skillsDir) continue;
        try {
          await installToAgent(skill.paths[0], agent.skillsDir, true);
          totalSuccess++;
        } catch {
          totalFail++;
        }
      }
    }
    const skillCount = selectedSkillsForInstall.length;
    const agentCount = selectedAgents.size;
    const msg = totalFail === 0
      ? `成功安装 ${skillCount} 个技能到 ${agentCount} 个智能体`
      : `完成：${totalSuccess} 成功，${totalFail} 失败`;
    setActionMsg({ success: totalFail === 0, msg: `安装：${msg}` });
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
    setSelectedAgents(new Set());
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
    if (selectedSkillsForInstall.length === 0 || !projectInstallPath.trim()) return;
    setInstalling(true);
    let success = 0, fail = 0;
    for (const skill of selectedSkillsForInstall) {
      try {
        await installToProject(skill.paths[0], projectInstallPath.trim(), true);
        success++;
      } catch {
        fail++;
      }
    }
    const msg = fail === 0
      ? `成功安装 ${success} 个技能到项目`
      : `完成：${success} 成功，${fail} 失败`;
    setActionMsg({ success: fail === 0, msg: `安装：${msg}` });
    setInstallDialogOpen(false);
    setSelectedSkillsForInstall([]);
    setSelectedPoolSkills(new Set());
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
    setActionMsg(null);
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

  const handleImport = async (skill: DiscoveredSkill) => {
    setImporting(true);
    setActionMsg(null);
    try {
      await importToPool(skill.path);
      setActionMsg({ success: true, msg: `"${skill.name}" 已复制到技能池` });
      setScanResults(prev => prev.filter(s => s.path !== skill.path));
      onRefresh();
    } catch (e: any) {
      setActionMsg({ success: false, msg: `复制失败: ${e.message}` });
    } finally {
      setImporting(false);
    }
  };

  const handleArchive = async (skill: DiscoveredSkill) => {
    setImporting(true);
    setActionMsg(null);
    try {
      await archiveToPool(skill.path);
      setActionMsg({ success: true, msg: `"${skill.name}" 已归档到技能池（原目录已移除）` });
      setScanResults(prev => prev.filter(s => s.path !== skill.path));
      onRefresh();
    } catch (e: any) {
      setActionMsg({ success: false, msg: `归档失败: ${e.message}` });
    } finally {
      setImporting(false);
    }
  };

  // Multi-select
  const toggleSelect = (path: string) => {
    setSelectedPaths(prev => {
      const next = new Set(prev);
      next.has(path) ? next.delete(path) : next.add(path);
      return next;
    });
  };

  const toggleSelectAll = () => {
    const allPaths = scanNew.map(s => s.path);
    if (selectedPaths.size === allPaths.length && allPaths.length > 0) {
      setSelectedPaths(new Set());
    } else {
      setSelectedPaths(new Set(allPaths));
    }
  };

  const handleImportToListed = async (skill: ListedSkill) => {
    if (skill.paths.length === 0) return;
    setImporting(true);
    setActionMsg(null);
    try {
      await importToPool(skill.paths[0]);
      setActionMsg({ success: true, msg: `"${skill.name}" 已归入技能池` });
      onRefresh();
    } catch (e: any) {
      setActionMsg({ success: false, msg: `入池失败: ${e.message}` });
    } finally {
      setImporting(false);
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
      if (!skill || skill.paths.length === 0) { f++; continue; }
      try { await importToPool(skill.paths[0]); s++; } catch { f++; }
    }
    setActionMsg({ success: f === 0, msg: `批量入池完成：${s} 个成功${f > 0 ? `，${f} 个失败` : ""}` });
    setSelectedGlobalSkills(new Set());
    onRefresh();
    setImporting(false);
  };

  const handleBatchImport = async () => {
    if (selectedPaths.size === 0) return;
    setImporting(true);
    let s = 0, f = 0;
    for (const path of selectedPaths) {
      try { await importToPool(path); s++; } catch { f++; }
    }
    setActionMsg({ success: f === 0, msg: `批量入库完成：${s} 个成功${f > 0 ? `，${f} 个失败` : ""}` });
    setSelectedPaths(new Set());
    onRefresh();
    setImporting(false);
  };

  const handleBatchArchive = async () => {
    if (selectedPaths.size === 0) return;
    setImporting(true);
    let s = 0, f = 0;
    for (const path of selectedPaths) {
      try { await archiveToPool(path); s++; } catch { f++; }
    }
    setActionMsg({ success: f === 0, msg: `批量归档完成：${s} 个成功${f > 0 ? `，${f} 个失败` : ""}` });
    setSelectedPaths(new Set());
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

  // Filtered global skills (all skills, match name/agent/path)
  const filteredGlobalSkills = globalKeyword.trim()
    ? skills.filter(s => {
        const kw = globalKeyword.toLowerCase();
        return (
          s.name.toLowerCase().includes(kw) ||
          s.agentNames.some(n => n.toLowerCase().includes(kw)) ||
          s.paths.some(p => p.toLowerCase().includes(kw))
        );
      })
    : skills;

  // Scan results split
  const scanInPool = scanResults.filter(s => s.alreadyInPool);
  const scanNew = scanResults.filter(s => !s.alreadyInPool);

  const allScanSelected = scanNew.length > 0 && scanNew.every(s => selectedPaths.has(s.path));

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">技能池</h2>
        <Button variant="outline" size="sm" onClick={onRefresh}>
          <RefreshCw className="h-4 w-4" /> 刷新
        </Button>
      </div>

      {actionMsg && (
        <Card>
          <CardContent className="py-3">
            <p className={`text-sm flex items-center gap-2 ${actionMsg.success ? "text-green-600" : "text-destructive"}`}>
              {actionMsg.success ? <CheckCircle2 className="h-4 w-4" /> : null}
              {actionMsg.msg}
            </p>
          </CardContent>
        </Card>
      )}

      <Tabs defaultValue="pool">
        <TabsList>
          <TabsTrigger value="pool">本机技能池 ({poolKeyword.trim() ? `${filteredPoolSkills.length}/${poolSkillsFromIndex.length}` : poolSkillsFromIndex.length})</TabsTrigger>
          <TabsTrigger value="global">智能体全局技能 ({skills.length})</TabsTrigger>
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
                      {filteredPoolSkills.length > 0 && selectedPoolSkills.size === filteredPoolSkills.length ? <CheckCircle2 className="h-3.5 w-3.5" /> : <span className="h-3.5 w-3.5 block">☐</span>}
                      {filteredPoolSkills.length > 0 && selectedPoolSkills.size === filteredPoolSkills.length ? "取消全选" : "全选"}
                    </Button>
                  </div>
                </div>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <input
                    className="w-full pl-9 pr-3 py-2 rounded-md border border-input bg-white text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
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
                            {filteredPoolSkills.length > 0 && selectedPoolSkills.size === filteredPoolSkills.length ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <span className="h-4 w-4 block leading-4 text-center text-muted-foreground">☐</span>}
                          </button>
                        </th>
                        <th className="text-left font-medium px-4 py-2 w-8"></th>
                        <th className="text-left font-medium px-4 py-2">技能名称</th>
                        <th className="text-left font-medium px-4 py-2">池路径</th>
                        <th className="text-right font-medium px-4 py-2">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredPoolSkills.map((s, idx) => {
                        const isSelected = selectedPoolSkills.has(s.name);
                        const poolSkillPath = poolPath ? `${poolPath}/${s.name}` : "";
                        return (
                        <tr key={s.name} className={`border-b last:border-0 hover:bg-muted/30 ${idx % 2 === 1 ? "bg-muted/[0.03]" : ""}`} onClick={() => onSelect(s)}>
                          <td className="px-4 py-2" onClick={(e) => e.stopPropagation()}>
                            <button
                              className="p-0.5 rounded hover:bg-muted"
                              onClick={() => toggleSelectPoolSkill(s.name)}
                            >
                              {isSelected ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <span className="h-4 w-4 block leading-4 text-center text-muted-foreground">☐</span>}
                            </button>
                          </td>
                          <td className="px-4 py-2">
                            <Badge variant="default" className="text-[10px]">池</Badge>
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
                              <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={(e) => { e.stopPropagation(); onSelect(s); }}>详情</Button>
                              <Button variant="outline" size="sm" className="h-7 text-xs" onClick={(e) => { e.stopPropagation(); openInstallDialog(s); }} disabled={installing}>
                                {installing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
                                安装
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
                      {filteredGlobalSkills.length > 0 && selectedGlobalSkills.size === filteredGlobalSkills.length ? <CheckCircle2 className="h-3.5 w-3.5" /> : <span className="h-3.5 w-3.5 block">☐</span>}
                      {filteredGlobalSkills.length > 0 && selectedGlobalSkills.size === filteredGlobalSkills.length ? "取消全选" : "全选"}
                    </Button>
                  </div>
                </div>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <input
                    className="w-full pl-9 pr-3 py-2 rounded-md border border-input bg-white text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    placeholder="筛选技能名称、智能体或路径..."
                    value={globalKeyword}
                    onChange={(e) => { setGlobalKeyword(e.target.value); setSelectedGlobalSkills(new Set()); }}
                  />
                </div>
                {globalKeyword.trim() && (
                  <p className="text-xs text-muted-foreground">
                    筛选 "{globalKeyword}"：{filteredGlobalSkills.length} / {skills.length} 个技能
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
                            {filteredGlobalSkills.length > 0 && selectedGlobalSkills.size === filteredGlobalSkills.length ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <span className="h-4 w-4 block leading-4 text-center text-muted-foreground">☐</span>}
                          </button>
                        </th>
                        <th className="text-left font-medium px-4 py-2">技能名称</th>
                        <th className="text-left font-medium px-4 py-2">智能体</th>
                        <th className="text-left font-medium px-4 py-2">池路径</th>
                        <th className="text-left font-medium px-4 py-2">状态</th>
                        <th className="text-right font-medium px-4 py-2">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredGlobalSkills.map((s, idx) => {
                        const poolSkillPath = poolPath && s.inPool ? `${poolPath}/${s.name}` : "";
                        return (
                        <tr key={s.name} className={`border-b last:border-0 hover:bg-muted/30 ${idx % 2 === 1 ? "bg-muted/[0.03]" : ""}`}>
                          <td className="px-4 py-2">
                            <button
                              className="p-0.5 rounded hover:bg-muted"
                              onClick={() => toggleSelectGlobal(s.name)}
                            >
                              {selectedGlobalSkills.has(s.name) ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <span className="h-4 w-4 block leading-4 text-center text-muted-foreground">☐</span>}
                            </button>
                          </td>
                          <td className="px-4 py-2 font-medium">{s.name}</td>
                          <td className="px-4 py-2">
                            <div className="flex flex-wrap gap-1">
                              {s.agentNames.map((name, i) => (
                                <Badge key={i} variant="secondary" className="text-[10px]">{name}</Badge>
                              ))}
                              {s.agentNames.length === 0 && <span className="text-xs text-muted-foreground">-</span>}
                            </div>
                          </td>
                          <td className="px-4 py-2">
                            {poolSkillPath ? (
                              <div className="flex items-center gap-1">
                                <span className="text-xs text-muted-foreground font-mono truncate max-w-[180px]" title={poolSkillPath}>{poolSkillPath}</span>
                                <Button variant="ghost" size="icon" className="h-5 w-5 shrink-0" onClick={(e) => { e.stopPropagation(); handleOpenFolder(poolSkillPath); }} disabled={openingFolder === poolSkillPath}>
                                  {openingFolder === poolSkillPath ? <Loader2 className="h-3 w-3 animate-spin" /> : <ExternalLink className="h-3 w-3" />}
                                </Button>
                              </div>
                            ) : <span className="text-xs text-muted-foreground">-</span>}
                          </td>
                          <td className="px-4 py-2">
                            {s.inPool ? (
                              <Badge variant="default" className="text-[10px]">在池中</Badge>
                            ) : (
                              <Badge variant="outline" className="text-[10px] text-muted-foreground">未入池</Badge>
                            )}
                          </td>
                          <td className="px-4 py-2 text-right">
                            <div className="flex justify-end gap-1">
                              <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={(e) => { e.stopPropagation(); onSelect(s); }}>详情</Button>
                              {s.inPool ? (
                                <Button variant="outline" size="sm" className="h-7 text-xs" onClick={(e) => { e.stopPropagation(); openInstallDialog(s); }} disabled={installing}>
                                  {installing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
                                  安装
                                </Button>
                              ) : (
                                <Button variant="outline" size="sm" className="h-7 text-xs" onClick={(e) => { e.stopPropagation(); handleImportToListed(s); }} disabled={importing || s.paths.length === 0}>
                                  {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Download className="h-3.5 w-3.5" />}
                                  归入技能池
                                </Button>
                              )}
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

        {/* Tab 3: Scan Results */}
        <TabsContent value="scan">
          {/* Scan Input */}
          <Card>
            <CardHeader>
              <CardTitle>本机扫描</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-2">
                <input
                  className="flex-1 px-3 py-2 rounded-md border border-input bg-background text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  placeholder="选择要扫描的目录（可选，留空则扫描所有智能体全局目录）"
                  value={projectPath}
                  onChange={(e) => setProjectPath(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleScan()}
                />
                <Button variant="outline" size="sm" onClick={handleBrowseProjectPath}>
                  <FolderOpen className="h-3.5 w-3.5" /> 浏览
                </Button>
                <Button onClick={handleScan} disabled={scanning}>
                  {scanning ? <Loader2 className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
                  扫描
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

          {/* Already in pool from scan */}
          {scanInPool.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-sm">
                  <CheckCircle2 className="h-4 w-4 text-green-500" />
                  已在池中 ({scanInPool.length})
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                  {scanInPool.map((s, i) => (
                    <div key={i} className={`border rounded-lg p-3 ${i % 2 === 1 ? "bg-muted/[0.03]" : ""}`}>
                      <div className="font-medium text-sm flex items-center gap-2">
                        <Puzzle className="h-3.5 w-3.5 text-primary" />
                        {s.name}
                        <Badge variant="default" className="text-[10px]">池</Badge>
                      </div>
                      <p className="text-xs text-muted-foreground mt-1 truncate" title={s.path}>{s.path}</p>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* New discoveries */}
          {scanNew.length > 0 && (
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center gap-2 text-sm">
                    <Search className="h-4 w-4 text-blue-500" />
                    新发现 ({scanNew.length})
                  </CardTitle>
                  {scanNew.length > 0 && (
                    <div className="flex items-center gap-2">
                      {selectedPaths.size > 0 && (
                        <>
                          <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleBatchImport} disabled={importing}>
                            {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Download className="h-3 w-3" />}
                            复制到池 ({selectedPaths.size})
                          </Button>
                          <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleBatchArchive} disabled={importing}>
                            {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Archive className="h-3 w-3" />}
                            归档到池 ({selectedPaths.size})
                          </Button>
                        </>
                      )}
                      <Button variant="ghost" size="sm" className="h-7 text-xs gap-1" onClick={toggleSelectAll}>
                        {allScanSelected ? <CheckCircle2 className="h-3.5 w-3.5" /> : <span className="h-3.5 w-3.5 block">☐</span>}
                        {allScanSelected ? "取消全选" : "全选"}
                      </Button>
                    </div>
                  )}
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                  {scanNew.map((s, i) => {
                    const isSelected = selectedPaths.has(s.path);
                    return (
                      <div key={i} className="border border-blue-200 bg-blue-50/20 rounded-lg p-3 space-y-2">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2 min-w-0">
                            <button
                              className="p-1 rounded hover:bg-muted shrink-0"
                              onClick={() => toggleSelect(s.path)}
                            >
                              {isSelected ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <span className="h-4 w-4 block leading-4 text-center text-muted-foreground">☐</span>}
                            </button>
                            <Puzzle className="h-3.5 w-3.5 text-blue-500 shrink-0" />
                            <span className="truncate">{s.name}</span>
                          </div>
                        </div>
                        <p className="text-xs text-muted-foreground font-mono truncate" title={s.path}>{s.path}</p>
                        {s.agentName && (
                          <Badge variant="outline" className="text-[10px]">{s.agentName}</Badge>
                        )}
                        <div className="flex gap-2">
                          <Button variant="outline" size="sm" className="h-7 text-xs flex-1" onClick={() => handleImport(s)} disabled={importing}>
                            {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Download className="h-3 w-3" />}
                            复制
                          </Button>
                          <Button variant="outline" size="sm" className="h-7 text-xs flex-1" onClick={() => handleArchive(s)} disabled={importing}>
                            {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Archive className="h-3 w-3" />}
                            归档
                          </Button>
                        </div>
                      </div>
                    );
                  })}
                </div>
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
                  {installing ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
                  确认安装
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
                  <input
                    className="flex-1 px-3 py-2 rounded-md border border-input bg-white text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
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
                <p className="text-xs text-muted-foreground mt-1">技能将安装到项目路径/.opencode/skills/ 目录</p>
              </div>
              {projectInstallPath.trim() && (
                <p className="text-xs text-muted-foreground">
                  将安装 {selectedSkillsForInstall.length} 个技能到项目路径（目录存在则覆盖）
                </p>
              )}
              <div className="flex justify-end gap-2 pt-2 border-t">
                <Button variant="outline" onClick={() => setInstallDialogOpen(false)} disabled={installing}>取消</Button>
                <Button onClick={handleInstallToProject} disabled={installing || !projectInstallPath.trim()}>
                  {installing ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
                  确认安装
                </Button>
              </div>
            </TabsContent>
          </Tabs>
        </DialogContent>
      </Dialog>
    </div>
  );
}
