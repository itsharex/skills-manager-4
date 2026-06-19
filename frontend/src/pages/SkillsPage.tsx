import { useState, useEffect } from "react";
import { RefreshCw, Search, Loader2, Puzzle, CheckCircle2, FolderOpen, Download, Archive, CheckSquare, Square } from "lucide-react";
import { toast } from "sonner";
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { scanLocal, listPool, getConfig, importToPool, archiveToPool, openDirectoryDialog } from "../bridge";
import type { ListedSkill, DiscoveredSkill } from "../types";

interface Props {
  skills: ListedSkill[];
  onSelect: (skill: ListedSkill) => void;
  onRefresh: () => void;
}

export default function SkillsPage({ skills, onSelect, onRefresh }: Props) {
  // Pool state
  const [poolSkills, setPoolSkills] = useState<DiscoveredSkill[]>([]);
  const [poolPath, setPoolPath] = useState("");
  const [loadingPool, setLoadingPool] = useState(true);

  // Scan state
  const [projectPath, setProjectPath] = useState("");
  const [scanning, setScanning] = useState(false);
  const [scanResults, setScanResults] = useState<DiscoveredSkill[]>([]);
  const [scanError, setScanError] = useState<string | null>(null);

  // Action state
  const [importing, setImporting] = useState(false);

  // Multi-select state for "not in pool" skills
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set());

  const inPoolSkills = skills.filter(s => s.inPool);
  const notInPoolSkills = skills.filter(s => !s.inPool);

  useEffect(() => {
    async function load() {
      try {
        const cfg = await getConfig();
        setPoolPath(cfg.pool_path || "");
        if (cfg.pool_path) {
          const pool = await listPool();
          setPoolSkills(pool);
        }
      } catch { /* ignore */ }
      finally { setLoadingPool(false); }
    }
    load();
  }, []);

  const reloadPool = async () => {
    try {
      const pool = await listPool();
      setPoolSkills(pool);
    } catch { /* ignore */ }
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

  const handleImport = async (skill: DiscoveredSkill) => {
    setImporting(true);
    try {
      await importToPool(skill.path);
      toast.success(`"${skill.name}" 已复制到技能池`);
      await reloadPool();
      setScanResults(prev => prev.filter(s => s.path !== skill.path));
      onRefresh();
    } catch (e: any) {
      toast.error(`复制失败: ${e.message}`);
    } finally {
      setImporting(false);
    }
  };

  const handleArchive = async (skill: DiscoveredSkill) => {
    setImporting(true);
    try {
      await archiveToPool(skill.path);
      toast.success(`"${skill.name}" 已归档到技能池（原目录已移除）`);
      await reloadPool();
      setScanResults(prev => prev.filter(s => s.path !== skill.path));
      onRefresh();
    } catch (e: any) {
      toast.error(`归档失败: ${e.message}`);
    } finally {
      setImporting(false);
    }
  };

  // Multi-select: select/deselect all not-in-pool skills
  const toggleSelectAll = () => {
    const allPaths = notInPoolSkills.flatMap(s => s.paths);
    if (selectedPaths.size === allPaths.length && allPaths.length > 0) {
      setSelectedPaths(new Set());
    } else {
      setSelectedPaths(new Set(allPaths));
    }
  };

  // Batch import: import all selected paths to pool (skip already in pool)
  const handleBatchImport = async () => {
    if (selectedPaths.size === 0) return;
    setImporting(true);
    let successCount = 0;
    let failCount = 0;
    for (const path of selectedPaths) {
      try {
        await importToPool(path);
        successCount++;
      } catch {
        failCount++;
      }
    }
    toast[failCount === 0 ? "success" : "error"](`批量入库完成：${successCount} 个成功${failCount > 0 ? `，${failCount} 个失败` : ""}`);
    setSelectedPaths(new Set());
    await reloadPool();
    onRefresh();
    setImporting(false);
  };

  // Batch archive: archive all selected paths to pool
  const handleBatchArchive = async () => {
    if (selectedPaths.size === 0) return;
    setImporting(true);
    let successCount = 0;
    let failCount = 0;
    for (const path of selectedPaths) {
      try {
        await archiveToPool(path);
        successCount++;
      } catch {
        failCount++;
      }
    }
    toast[failCount === 0 ? "success" : "error"](`批量归档完成：${successCount} 个成功${failCount > 0 ? `，${failCount} 个失败` : ""}`);
    setSelectedPaths(new Set());
    await reloadPool();
    onRefresh();
    setImporting(false);
  };

  const scanInPool = scanResults.filter(s => s.alreadyInPool);
  const scanNew = scanResults.filter(s => !s.alreadyInPool);

  const allNotInPoolPaths = notInPoolSkills.flatMap(s => s.paths);
  const allSelected = allNotInPoolPaths.length > 0 && selectedPaths.size === allNotInPoolPaths.length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">技能管理</h2>
          <p className="text-muted-foreground mt-1">
            技能池: {poolPath || "未配置"} {poolPath && !loadingPool && <span className="text-xs">({poolSkills.length} 个)</span>}
            {" | "}本机共 {skills.length} 个技能
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={onRefresh}>
          <RefreshCw className="h-4 w-4" /> 刷新
        </Button>
      </div>

      {/* Section 1: Pool Skills Summary */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-sm">
            <CheckCircle2 className="h-4 w-4 text-green-500" />
            技能池汇总 ({inPoolSkills.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {inPoolSkills.length === 0 ? (
            <p className="text-muted-foreground text-sm py-4 text-center">技能池为空</p>
          ) : (
            <div className="border rounded-lg overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-muted/50 border-b">
                    <th className="text-left font-medium px-4 py-2">技能名称</th>
                    <th className="text-left font-medium px-4 py-2">智能体工具</th>
                    <th className="text-right font-medium px-4 py-2"></th>
                  </tr>
                </thead>
                <tbody>
                  {inPoolSkills.map((s) => (
                    <tr key={s.name} className={`border-b last:border-0 hover:bg-muted/30 cursor-pointer`} onClick={() => onSelect(s)}>
                      <td className="px-4 py-2 font-medium flex items-center gap-2">
                        <Puzzle className="h-3.5 w-3.5 text-primary shrink-0" />
                        {s.name}
                        <Badge variant="default" className="text-[10px] ml-1">池</Badge>
                      </td>
                      <td className="px-4 py-2">
                        <div className="flex flex-wrap gap-1">
                          {s.agentNames.map((name, i) => (
                            <Badge key={i} variant="outline" className="text-[10px]">{name}</Badge>
                          ))}
                        </div>
                      </td>
                      <td className="px-4 py-2 text-right">
                        <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); onSelect(s); }}>详情</Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Section 2: All Installed Skills (not in pool) — with multi-select */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Puzzle className="h-4 w-4 text-blue-500" />
              本机技能汇总 ({notInPoolSkills.length})
            </CardTitle>
            {notInPoolSkills.length > 0 && (
              <div className="flex items-center gap-2">
                {selectedPaths.size > 0 && (
                  <>
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 text-xs"
                      onClick={handleBatchImport}
                      disabled={importing}
                    >
                      {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Download className="h-3 w-3" />}
                      复制到池 ({selectedPaths.size})
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 text-xs"
                      onClick={handleBatchArchive}
                      disabled={importing}
                    >
                      {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Archive className="h-3 w-3" />}
                      归档到池 ({selectedPaths.size})
                    </Button>
                  </>
                )}
                <Button variant="ghost" size="sm" className="h-7 text-xs gap-1" onClick={toggleSelectAll}>
                  {allSelected ? <CheckSquare className="h-3.5 w-3.5" /> : <Square className="h-3.5 w-3.5" />}
                  {allSelected ? "取消全选" : "全选"}
                </Button>
              </div>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {notInPoolSkills.length === 0 ? (
            <p className="text-muted-foreground text-sm py-4 text-center">无额外技能</p>
          ) : (
            <div className="border rounded-lg overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-muted/50 border-b">
                    <th className="w-8 px-2 py-2"></th>
                    <th className="text-left font-medium px-4 py-2">技能名称</th>
                    <th className="text-left font-medium px-4 py-2">智能体工具</th>
                    <th className="text-right font-medium px-4 py-2"></th>
                  </tr>
                </thead>
                <tbody>
                  {notInPoolSkills.map((s) => {
                    // A skill is "selected" if ALL its paths are selected
                    const allPathsSelected = s.paths.length > 0 && s.paths.every(p => selectedPaths.has(p));
                    const somePathsSelected = s.paths.some(p => selectedPaths.has(p));
                    return (
                      <tr
                        key={s.name}
                        className={`border-b last:border-0 hover:bg-muted/30 cursor-pointer ${somePathsSelected ? "bg-primary/5" : ""}`}
                        onClick={() => onSelect(s)}
                      >
                        <td className="px-2 py-2" onClick={(e) => e.stopPropagation()}>
                          <button
                            className="p-0.5 rounded hover:bg-muted"
                            onClick={() => {
                              // Toggle all paths for this skill
                              setSelectedPaths(prev => {
                                const next = new Set(prev);
                                if (allPathsSelected) {
                                  s.paths.forEach(p => next.delete(p));
                                } else {
                                  s.paths.forEach(p => next.add(p));
                                }
                                return next;
                              });
                            }}
                          >
                            {allPathsSelected
                              ? <CheckSquare className="h-4 w-4 text-primary" />
                              : somePathsSelected
                                ? <CheckSquare className="h-4 w-4 text-primary/50" />
                                : <Square className="h-4 w-4 text-muted-foreground" />
                            }
                          </button>
                        </td>
                        <td className="px-4 py-2 font-medium flex items-center gap-2">
                          <Puzzle className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                          {s.name}
                        </td>
                        <td className="px-4 py-2">
                          <div className="flex flex-wrap gap-1">
                            {s.agentNames.map((name, i) => (
                              <Badge key={i} variant="outline" className="text-[10px]">{name}</Badge>
                            ))}
                          </div>
                        </td>
                        <td className="px-4 py-2 text-right">
                          <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); onSelect(s); }}>详情</Button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Section 3: Local Scan */}
      <Card>
        <CardHeader>
          <CardTitle>本机扫描</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-xs text-muted-foreground">
            递归扫描指定目录下所有子目录，寻找包含 SKILL.md 的技能定义。
          </p>
          <div className="flex gap-2">
            <Input
              className="flex-1"
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

      {/* Scan Results: Already in pool */}
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
                <div key={i} className="border rounded-lg p-3 bg-muted/30">
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

      {/* Scan Results: New discoveries */}
      {scanNew.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-sm">
              <Search className="h-4 w-4 text-blue-500" />
              新发现 ({scanNew.length})
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {scanNew.map((s, i) => (
                <div key={i} className="border rounded-lg p-3 border-blue-200 bg-blue-50/20 space-y-2">
                  <div className="font-medium text-sm flex items-center gap-2">
                    <Puzzle className="h-3.5 w-3.5 text-blue-500 shrink-0" />
                    <span className="truncate">{s.name}</span>
                  </div>
                  <p className="text-xs text-muted-foreground font-mono truncate" title={s.path}>{s.path}</p>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 text-xs flex-1"
                      onClick={() => handleImport(s)}
                      disabled={importing}
                    >
                      {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Download className="h-3 w-3" />}
                      复制到池
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 text-xs flex-1"
                      onClick={() => handleArchive(s)}
                      disabled={importing}
                    >
                      {importing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Archive className="h-3 w-3" />}
                      归档到池
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
