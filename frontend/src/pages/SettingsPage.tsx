import { useState, useEffect } from "react";
import { RefreshCw, ChevronDown, ChevronUp, FolderOpen, Plus, Trash2, Globe, Database, ToggleLeft, ToggleRight, AlertCircle, Key, ScrollText } from "lucide-react";
import { toast } from "sonner";
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { useI18n } from "../i18n/context";
import { getConfig, saveConfig, openDirectoryDialog, getOpLogs } from "../bridge";
import type { AgentGroup, AgentInfo, Config, MarketSource, OpLog } from "../types";

interface Props {
  agents: AgentInfo[];
  onRefresh: () => void;
}

function groupAgentsByPath(agents: AgentInfo[]): AgentGroup[] {
  const groups = new Map<string, AgentGroup>();

  const detected = agents.filter(a => a.detected);
  const undetected = agents.filter(a => !a.detected);

  const allGrouped = [...detected, ...undetected];
  for (const a of allGrouped) {
    const key = a.path;
    if (!groups.has(key)) {
      groups.set(key, {
        path: key,
        agents: [{ id: a.id, name: a.name }],
        detected: a.detected,
        displayName: a.name,
        tooltipName: a.name,
      });
    } else {
      const group = groups.get(key)!;
      group.agents.push({ id: a.id, name: a.name });
      if (a.detected) group.detected = true;
    }
  }

  for (const group of groups.values()) {
    const sorted = [...group.agents].sort((a, b) => a.name.length - b.name.length);
    const shortestName = sorted[0]?.name ?? group.agents[0]?.name ?? "";
    group.displayName = group.detected ? shortestName : `${shortestName}等`;
    group.tooltipName = group.agents.map(a => a.name).join(", ");
  }

  return Array.from(groups.values());
}

const MARKET_TYPE_ICONS: Record<string, React.ReactNode> = {
  pool: <Database className="h-3.5 w-3.5" />,
  github: <Globe className="h-3.5 w-3.5" />,
  registry: <Globe className="h-3.5 w-3.5" />,
};

function validateMarketSource(source: Partial<MarketSource>): string | null {
  if (!source.name?.trim()) return "来源名称不能为空";
  if (!source.url?.trim()) return "来源 URL 不能为空";
  if (!source.type) return "来源类型不能为空";
  if (source.type === "github") {
    const url = source.url.trim();
    if (!url.includes("github.com/") && !url.startsWith("gh:")) {
      return "GitHub 来源必须包含 github.com/owner/repo 格式的 URL";
    }
  }
  if (source.type === "pool") {
    if (!source.url.startsWith("/") && !source.url.startsWith("~") && !source.url.startsWith(".")) {
      return "本地池来源 URL 必须为本地路径";
    }
  }
  return null;
}

export default function SettingsPage({ agents, onRefresh }: Props) {
  const { t } = useI18n();
  const [showAll, setShowAll] = useState(false);
  const INITIAL_SHOW = 10;

  // Config state
  const [config, setConfig] = useState<Config | null>(null);
  const [poolPath, setPoolPath] = useState("");
  const [savingPoolPath, setSavingPoolPath] = useState(false);
  const [gitHubToken, setGitHubToken] = useState("");
  const [savingToken, setSavingToken] = useState(false);

  // Market sources state
  const [marketSources, setMarketSources] = useState<MarketSource[]>([]);
  const [newSource, setNewSource] = useState<Partial<MarketSource>>({ name: "", url: "", type: "github", enabled: true, branch: "main" });
  const [sourceValidation, setSourceValidation] = useState<string | null>(null);
  const [savingSources, setSavingSources] = useState(false);

  // Operation logs state
  const [opLogs, setOpLogs] = useState<OpLog[]>([]);
  const [loadingLogs, setLoadingLogs] = useState(false);
  const [showLogs, setShowLogs] = useState(false);

  // Load config on mount
  useEffect(() => {
    async function load() {
      try {
        const cfg = await getConfig();
        setConfig(cfg);
        setPoolPath(cfg.pool_path || "");
        setGitHubToken(cfg.github_token || "");
        setMarketSources(cfg.market_sources || []);
      } catch {
        // backend not available
      }
    }
    load();
  }, []);

  const handleSavePoolPath = async () => {
    if (!config) return;
    setSavingPoolPath(true);
    try {
      await saveConfig({ ...config, pool_path: poolPath });
      toast.success("技能池路径已保存");
    } catch (e: any) {
      toast.error(`保存失败: ${e.message}`);
    } finally {
      setSavingPoolPath(false);
    }
  };

  const handleBrowsePoolPath = async () => {
    try {
      const dir = await openDirectoryDialog("选择技能池目录");
      if (dir) {
        setPoolPath(dir);
      }
    } catch (e: any) {
      // User canceled dialog
    }
  };

  const handleSaveGitHubToken = async () => {
    if (!config) return;
    setSavingToken(true);
    try {
      await saveConfig({ ...config, github_token: gitHubToken.trim() });
      toast.success(gitHubToken.trim() ? "GitHub Token 已保存" : "GitHub Token 已清除");
    } catch (e: any) {
      toast.error(`保存失败: ${e.message}`);
    } finally {
      setSavingToken(false);
    }
  };

  const handleBrowseSourceUrl = async () => {
    try {
      const dir = await openDirectoryDialog("选择本地池目录");
      if (dir) {
        setNewSource({ ...newSource, url: dir });
        setSourceValidation(null);
      }
    } catch (e: any) {
      // User canceled dialog
    }
  };

  const handleAddSource = async () => {
    const err = validateMarketSource(newSource);
    setSourceValidation(err);
    if (err) return;

    const source: MarketSource = {
      name: newSource.name!.trim(),
      url: newSource.url!.trim(),
      type: newSource.type || "github",
      enabled: newSource.enabled !== false,
      branch: newSource.type === "github" ? (newSource.branch || "main") : undefined,
    };
    const updated = [...marketSources, source];
    setMarketSources(updated);
    setNewSource({ name: "", url: "", type: "github", enabled: true, branch: "main" });
    setSourceValidation(null);

    // Persist
    if (config) {
      setSavingSources(true);
      try {
        await saveConfig({ ...config, market_sources: updated });
        toast.success("市场来源已添加");
      } catch (e: any) {
        toast.error(`保存失败: ${e.message}`);
      } finally {
        setSavingSources(false);
      }
    }
  };

  const handleToggleSource = async (index: number) => {
    const updated = marketSources.map((s, i) =>
      i === index ? { ...s, enabled: !s.enabled } : s
    );
    setMarketSources(updated);
    if (config) {
      try {
        await saveConfig({ ...config, market_sources: updated });
      } catch {
        // ignore
      }
    }
  };

  const handleRemoveSource = async (index: number) => {
    const updated = marketSources.filter((_, i) => i !== index);
    setMarketSources(updated);
    if (config) {
      try {
        await saveConfig({ ...config, market_sources: updated });
      } catch {
        // ignore
      }
    }
  };

  const handleLoadLogs = async () => {
    if (showLogs) {
      setShowLogs(false);
      return;
    }
    setLoadingLogs(true);
    try {
      const logs = await getOpLogs(50);
      setOpLogs(logs || []);
      setShowLogs(true);
    } catch {
      toast.error("加载操作日志失败");
    } finally {
      setLoadingLogs(false);
    }
  };

  const groups = groupAgentsByPath(agents);
  const displayed = showAll ? groups : groups.slice(0, INITIAL_SHOW);
  const remaining = groups.length - INITIAL_SHOW;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">{t("settings.title")}</h2>
        <p className="text-muted-foreground mt-1">{t("settings.subtitle")}</p>
      </div>

      {/* Pool Path Configuration */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FolderOpen className="h-4 w-4" />
            技能池路径
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-xs text-muted-foreground">
            配置本地技能池目录（默认 ~/.skill-pool/）。该目录作为本机技能仓库，扫描结果将与此目录交叉匹配。
          </p>
          <div className="flex gap-2">
            <Input
              className="flex-1"
              placeholder="~/.skill-pool/"
              value={poolPath}
              onChange={(e) => { setPoolPath(e.target.value); }}
            />
            <Button variant="outline" size="sm" onClick={handleBrowsePoolPath}>
              <FolderOpen className="h-3.5 w-3.5" />
              浏览
            </Button>
            <Button onClick={handleSavePoolPath} disabled={savingPoolPath || !poolPath.trim()}>
              {savingPoolPath ? "保存中..." : "保存"}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* GitHub Token Configuration */}
      <Card className="border-t pt-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Key className="h-4 w-4" />
            GitHub Token
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-xs text-muted-foreground">
            配置 GitHub Personal Access Token 以提升 API 访问配额。未配置时 GitHub API 限制为 60 次/小时，配置后可提升至 5000 次/小时。
            ClawHub 和 skills.sh 搜索均依赖 GitHub API，遇到 403 错误时请配置此 Token。
          </p>
          <div className="flex gap-2">
            <Input
              className="flex-1"
              type="password"
              placeholder="ghp_xxxxxxxxxxxxxxxxxxxx"
              value={gitHubToken}
              onChange={(e) => { setGitHubToken(e.target.value); }}
            />
            <Button onClick={handleSaveGitHubToken} disabled={savingToken}>
              {savingToken ? "保存中..." : "保存"}
            </Button>
          </div>
          <p className="text-[10px] text-muted-foreground">
            生成方式：GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic) → Generate new token（无需勾选任何权限）
          </p>
        </CardContent>
      </Card>

      {/* Market Sources Management */}
      <Card className="border-t pt-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Globe className="h-4 w-4" />
            市场来源
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-xs text-muted-foreground">
            配置技能市场搜索来源。支持本地池（pool）、GitHub 仓库（github）和开放市场（registry）。
            搜索时将按列表顺序依次搜索。
          </p>

          {/* Existing sources list */}
          {marketSources.length === 0 ? (
            <p className="text-sm text-muted-foreground py-2 text-center">暂无市场来源。请添加一个来源开始使用。</p>
          ) : (
            <div className="space-y-2">
              {marketSources.map((source, i) => (
                <div key={i} className="flex items-center justify-between border rounded-lg px-3 py-2">
                  <div className="flex items-center gap-3 min-w-0">
                    <span className="text-muted-foreground shrink-0">
                      {MARKET_TYPE_ICONS[source.type] || <Globe className="h-3.5 w-3.5" />}
                    </span>
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium truncate">{source.name}</span>
                        <Badge variant="outline" className="text-[10px]">{source.type}</Badge>
                        {!source.enabled && (
                          <Badge variant="secondary" className="text-[10px]">已禁用</Badge>
                        )}
                      </div>
                      <p className="text-xs text-muted-foreground truncate font-mono">{source.url}</p>
                      {source.branch && (
                        <p className="text-[10px] text-muted-foreground">分支: {source.branch}</p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-1 shrink-0">
                    <Button variant="ghost" size="sm" onClick={() => handleToggleSource(i)} title={source.enabled ? "禁用" : "启用"}>
                      {source.enabled ? <ToggleRight className="h-4 w-4 text-green-500" /> : <ToggleLeft className="h-4 w-4 text-muted-foreground" />}
                    </Button>
                    <Button variant="ghost" size="sm" onClick={() => handleRemoveSource(i)} title="删除">
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Add new source form */}
          <div className="border rounded-lg p-3 space-y-3">
            <p className="text-xs font-medium text-muted-foreground">添加新来源</p>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <Input
                placeholder="名称（如 my-repo）"
                value={newSource.name || ""}
                onChange={(e) => setNewSource({ ...newSource, name: e.target.value })}
              />
              <div className="flex gap-2">
                <Input
                  className="flex-1"
                  placeholder="URL（GitHub 地址或本地路径）"
                  value={newSource.url || ""}
                  onChange={(e) => setNewSource({ ...newSource, url: e.target.value })}
                />
                {(newSource.type === "pool") && (
                  <Button variant="outline" size="sm" onClick={handleBrowseSourceUrl}>
                    <FolderOpen className="h-3.5 w-3.5" />
                  </Button>
                )}
              </div>
              <select
                className="px-3 py-2 rounded-md border border-input bg-background text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                value={newSource.type || "github"}
                onChange={(e) => setNewSource({ ...newSource, type: e.target.value, url: "" })}
              >
                <option value="github">GitHub</option>
                <option value="pool">本地池 (pool)</option>
                <option value="registry">开放市场 (registry)</option>
              </select>
              {newSource.type === "github" && (
                <Input
                  placeholder="分支（默认 main）"
                  value={newSource.branch || "main"}
                  onChange={(e) => setNewSource({ ...newSource, branch: e.target.value })}
                />
              )}
            </div>
            {sourceValidation && (
              <p className="text-xs text-destructive flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {sourceValidation}
              </p>
            )}
            <div className="flex justify-end">
              <Button size="sm" onClick={handleAddSource} disabled={savingSources}>
                <Plus className="h-4 w-4 mr-1" />
                添加来源
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Agents Detection */}
      <Card className="border-t pt-6">
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>{t("settings.agents")}</CardTitle>
          <Button variant="outline" size="sm" onClick={onRefresh}>
            <RefreshCw className="h-4 w-4" />
            {t("settings.redetect")}
          </Button>
        </CardHeader>
        <CardContent>
          {groups.length === 0 ? (
            <p className="text-muted-foreground text-sm py-4 text-center">{t("settings.no_agents")}</p>
          ) : (
            <>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                {displayed.map((group) => {
                  return (
                    <div key={group.path} className={`border rounded-lg p-3 space-y-2 ${group.detected ? "border-green-200 bg-green-50/30" : ""}`}>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2 min-w-0">
                          <span className={`w-2 h-2 rounded-full shrink-0 ${group.detected ? "bg-green-500" : "bg-gray-300"}`} />
                          <span className="text-xs font-medium truncate" title={group.tooltipName}>
                            {group.displayName}
                          </span>
                        </div>
                      </div>
                      <p className="text-xs text-muted-foreground font-mono truncate" title={group.path}>
                        {group.path}
                      </p>
                      <div className="flex flex-wrap gap-1">
                        {group.agents.map(a => (
                          <span key={a.id} className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                            {a.id}
                          </span>
                        ))}
                      </div>
                      <Badge variant={group.detected ? "default" : "secondary"} className="text-[10px]">
                        {group.detected ? t("settings.detected") : t("settings.not_detected")}
                      </Badge>
                    </div>
                  );
                })}
              </div>
              {groups.length > INITIAL_SHOW && (
                <div className="mt-4 text-center">
                  <Button variant="outline" size="sm" onClick={() => setShowAll(!showAll)}>
                    {showAll ? (
                      <><ChevronUp className="h-4 w-4" /> Show less</>
                    ) : (
                      <><ChevronDown className="h-4 w-4" /> Show all ({remaining} more)</>
                    )}
                  </Button>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Operation Logs */}
      <Card className="border-t pt-6">
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <ScrollText className="h-4 w-4" />
            操作日志
          </CardTitle>
          <Button variant="outline" size="sm" onClick={handleLoadLogs} disabled={loadingLogs}>
            <RefreshCw className={`h-4 w-4 ${loadingLogs ? "animate-spin" : ""}`} />
            {showLogs ? "收起" : "查看日志"}
          </Button>
        </CardHeader>
        {showLogs && (
          <CardContent>
            {opLogs.length === 0 ? (
              <p className="text-sm text-muted-foreground py-4 text-center">暂无操作日志</p>
            ) : (
              <div className="space-y-1.5 max-h-96 overflow-y-auto">
                {opLogs.slice().reverse().map((log, i) => (
                  <div key={i} className={`text-xs font-mono px-2 py-1.5 rounded border ${log.success ? "border-green-100 bg-green-50/30" : "border-red-100 bg-red-50/30"}`}>
                    <div className="flex items-center gap-2">
                      <span className="text-muted-foreground shrink-0">{new Date(log.timestamp).toLocaleString("zh-CN")}</span>
                      <Badge variant={log.success ? "default" : "destructive"} className="text-[10px] px-1.5 py-0">
                        {log.operation}
                      </Badge>
                      <span className="font-medium">{log.target}</span>
                      {log.agents && <span className="text-muted-foreground">→ {log.agents}</span>}
                    </div>
                    <div className="text-muted-foreground mt-0.5">{log.detail}</div>
                    {log.storePath && <div className="text-blue-600 mt-0.5">存储: {log.storePath}</div>}
                    {log.source && <div className="text-muted-foreground mt-0.5">来源: {log.source}</div>}
                    {log.error && <div className="text-red-600 mt-0.5">错误: {log.error}</div>}
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        )}
      </Card>
    </div>
  );
}