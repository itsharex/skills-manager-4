import { useState, useCallback } from "react";
import { Search, Loader2, AlertCircle, Globe, Database, Download, Sparkles, ExternalLink } from "lucide-react";
import { toast } from "sonner";
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card";
import { Button } from "../components/ui/button";
import { Badge } from "../components/ui/badge";
import { Input } from "../components/ui/input";
import AgentSelector from "../components/AgentSelector";
import { useI18n } from "../i18n/context";
import { searchMarket, installMarketSkill, openURL } from "../bridge";
import type { MarketSearchResult, MarketSearchSkill } from "../types";

const SOURCE_CONFIG: Record<string, { label: string; icon: React.ReactNode; bgColor: string; borderColor: string; headerBg: string; rowBorder: string }> = {
  pool: {
    label: "本地技能池",
    icon: <Database className="h-4 w-4 text-blue-600" />,
    bgColor: "bg-blue-50 dark:bg-blue-950/30",
    borderColor: "border-blue-200 dark:border-blue-800",
    headerBg: "bg-blue-100 dark:bg-blue-900/40",
    rowBorder: "border-l-blue-500",
  },
  clawhub: {
    label: "ClawHub",
    icon: <Sparkles className="h-4 w-4 text-purple-600" />,
    bgColor: "bg-purple-50 dark:bg-purple-950/30",
    borderColor: "border-purple-200 dark:border-purple-800",
    headerBg: "bg-purple-100 dark:bg-purple-900/40",
    rowBorder: "border-l-purple-500",
  },
  skillssh: {
    label: "skills.sh",
    icon: <Globe className="h-4 w-4 text-emerald-600" />,
    bgColor: "bg-emerald-50 dark:bg-emerald-950/30",
    borderColor: "border-emerald-200 dark:border-emerald-800",
    headerBg: "bg-emerald-100 dark:bg-emerald-900/40",
    rowBorder: "border-l-emerald-500",
  },
  github: {
    label: "",
    icon: <Globe className="h-4 w-4 text-gray-600" />,
    bgColor: "bg-gray-50 dark:bg-gray-950/30",
    borderColor: "border-gray-200 dark:border-gray-800",
    headerBg: "bg-gray-100 dark:bg-gray-900/40",
    rowBorder: "border-l-gray-500",
  },
  registry: {
    label: "",
    icon: <Globe className="h-4 w-4 text-orange-600" />,
    bgColor: "bg-orange-50 dark:bg-orange-950/30",
    borderColor: "border-orange-200 dark:border-orange-800",
    headerBg: "bg-orange-100 dark:bg-orange-900/40",
    rowBorder: "border-l-orange-500",
  },
};

interface Props {
  onRefresh: () => void;
}

export default function MarketPage({ onRefresh }: Props) {
  const { t } = useI18n();
  const [keyword, setKeyword] = useState("");
  const [searching, setSearching] = useState(false);
  const [results, setResults] = useState<MarketSearchResult[]>([]);
  const [searchError, setSearchError] = useState<string | null>(null);
  const [selectedAgents, setSelectedAgents] = useState<string[]>([]);
  const [installing, setInstalling] = useState<string | null>(null);
  const [installLog, setInstallLog] = useState<string[]>([]);
  const [hasSearched, setHasSearched] = useState(false);

  const handleSearch = useCallback(async () => {
    const kw = keyword.trim();
    if (!kw) return;

    setSearching(true);
    setSearchError(null);
    setInstallLog([]);
    setHasSearched(true);

    try {
      const data = await searchMarket(kw);
      setResults(data);
    } catch (e: any) {
      setSearchError(typeof e === "string" ? e : (e?.message || "搜索失败"));
      setResults([]);
    } finally {
      setSearching(false);
    }
  }, [keyword]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSearch();
    }
  };

  const handleInstall = async (skill: MarketSearchSkill) => {
    if (selectedAgents.length === 0) {
      toast.error("请先在上方「安装目标 Agent」中选择目标智能体");
      return;
    }

    setInstalling(skill.name);
    try {
      const logs = await installMarketSkill(skill, selectedAgents);
      const newLogs = logs.map((r: any) =>
        r.error
          ? `[FAIL] ${r.skillName || skill.name}: ${r.error}`
          : `[OK] ${r.skillName || skill.name}@${r.version} -> ${r.path}`
      );
      setInstallLog(prev => [...prev, ...newLogs]);

      const failCount = logs.filter((r: any) => r.error).length;
      const okCount = logs.length - failCount;
      if (failCount === 0 && okCount > 0) {
        toast.success(`${skill.name} 安装成功 (${okCount} 个 Agent)`);
      } else if (okCount > 0 && failCount > 0) {
        toast.warning(`${skill.name}: ${okCount} 成功, ${failCount} 失败`);
      } else if (failCount > 0) {
        toast.error(`${skill.name} 安装失败`);
      }

      onRefresh();
    } catch (e: any) {
      // Wails v2: Go errors come as strings in rejected promises, not Error objects
      const errMsg = typeof e === "string" ? e : (e?.message || e?.toString?.() || String(e));
      setInstallLog(prev => [...prev, `[FAIL] ${skill.name}: ${errMsg}`]);
      toast.error(`安装失败: ${errMsg}`);
    } finally {
      setInstalling(null);
    }
  };

  // Build external link URL for a skill's source.
  const buildExternalUrl = (skill: MarketSearchSkill, sourceType: string): string | null => {
    if (!skill.source) return null;
    switch (sourceType) {
      case "skillssh":
        return `https://skills.sh/${skill.source}`;
      case "clawhub":
      case "github":
        return `https://github.com/${skill.source}`;
      case "registry":
        if (skill.source.startsWith("http")) return skill.source;
        return null;
      default:
        return null;
    }
  };

  const formatInstalls = (count: number): string => {
    if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M`;
    if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K`;
    return `${count}`;
  };

  const totalSkills = results.reduce((sum, r) => sum + (r.skills || []).length, 0);

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">{t("market.title")}</h2>
        <p className="text-muted-foreground mt-1">{t("market.subtitle")}</p>
      </div>

      {/* Search Bar */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex gap-3">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                className="pl-9"
                placeholder="搜索技能名称、描述或标签..."
                value={keyword}
                onChange={(e) => setKeyword(e.target.value)}
                onKeyDown={handleKeyDown}
                disabled={searching}
              />
            </div>
            <Button onClick={handleSearch} disabled={searching || !keyword.trim()}>
              {searching ? (
                <><Loader2 className="h-4 w-4 animate-spin" /> 搜索中...</>
              ) : (
                <><Search className="h-4 w-4" /> 搜索</>
              )}
            </Button>
          </div>
          <p className="text-xs text-muted-foreground mt-2">
            搜索范围：本地技能池、ClawHub、skills.sh
          </p>
        </CardContent>
      </Card>

      {/* Search Error */}
      {searchError && (
        <Card>
          <CardContent className="py-4">
            <p className="text-sm text-destructive flex items-center gap-2">
              <AlertCircle className="h-4 w-4" />
              {searchError}
            </p>
          </CardContent>
        </Card>
      )}

      {/* Agent Selector - always visible so users can select before searching */}
      <AgentSelector onSelectionChange={setSelectedAgents} />

      {/* Empty State */}
      {!hasSearched ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Globe className="h-10 w-10 mx-auto text-muted-foreground mb-3" />
            <p className="text-sm text-muted-foreground mb-1">
              输入关键词搜索技能
            </p>
            <p className="text-xs text-muted-foreground">
              支持搜索本地技能池、ClawHub 市场、skills.sh 以及配置的 GitHub 仓库
            </p>
          </CardContent>
        </Card>
      ) : searching ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Loader2 className="h-8 w-8 mx-auto animate-spin text-muted-foreground mb-3" />
            <p className="text-sm text-muted-foreground">正在搜索各来源...</p>
          </CardContent>
        </Card>
      ) : totalSkills === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Search className="h-10 w-10 mx-auto text-muted-foreground mb-3" />
            <p className="text-sm text-muted-foreground mb-1">
              未找到与 "{keyword}" 相关的技能
            </p>
            <p className="text-xs text-muted-foreground">
              尝试使用不同的关键词或检查网络连接
            </p>
          </CardContent>
        </Card>
      ) : null}

      {/* Results by Source - each source gets its own visually distinct card */}
      {hasSearched && !searching && results.map((result, ri) => {
        const cfg = SOURCE_CONFIG[result.sourceType] || SOURCE_CONFIG.registry;
        const headerLabel = result.sourceType === "github" || result.sourceType === "registry"
          ? result.sourceName
          : cfg.label;
        const skills = result.skills || [];

        if (skills.length === 0 && !result.error) return null;

        return (
          <Card key={ri} className={`${cfg.borderColor} overflow-hidden`}>
            {/* Source Header with colored background */}
            <div className={`${cfg.headerBg} px-6 py-3 flex items-center gap-2 border-b ${cfg.borderColor}`}>
              {cfg.icon}
              <span className="text-sm font-bold">{headerLabel}</span>
              <Badge variant="outline" className="text-[10px] border-current">
                {skills.length} 个
              </Badge>
              {result.error && (
                <span className="text-xs text-destructive flex items-center gap-1 ml-auto">
                  <AlertCircle className="h-3 w-3" />
                  {result.error}
                </span>
              )}
            </div>

            {/* Skills Table */}
            {skills.length > 0 && (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="bg-muted/50 border-b">
                      <th className="text-left font-medium px-4 py-2 w-[28%]">名称</th>
                      <th className="text-left font-medium px-4 py-2 w-[12%]">安装量</th>
                      <th className="text-left font-medium px-4 py-2 w-[22%]">来源</th>
                      <th className="text-left font-medium px-4 py-2 w-[25%]">描述</th>
                      <th className="text-right font-medium px-4 py-2 w-[13%]">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {skills.map((skill, si) => (
                      <tr key={si} className={`border-b last:border-0 hover:bg-muted/30 border-l-4 ${cfg.rowBorder}`}>
                        <td className="px-4 py-2">
                          <span className="font-medium">{skill.name}</span>
                        </td>
                        <td className="px-4 py-2">
                          <Badge variant="secondary" className="text-[10px]">
                            {skill.installs ? formatInstalls(skill.installs) : (skill.version || "-")}
                          </Badge>
                        </td>
                        <td className="px-4 py-2 text-muted-foreground font-mono text-xs">
                          {(() => {
                            const url = buildExternalUrl(skill, result.sourceType);
                            const label = skill.source || skill.namespace;
                            if (!url) return <span>{label}</span>;
                            return (
                              <button
                                className="text-blue-600 hover:text-blue-800 hover:underline inline-flex items-center gap-1 cursor-pointer"
                                onClick={(e) => { e.stopPropagation(); openURL(url); }}
                                title={`在浏览器中打开: ${url}`}
                              >
                                {label}
                                <ExternalLink className="h-3 w-3" />
                              </button>
                            );
                          })()}
                        </td>
                        <td className="px-4 py-2">
                          <span className="text-xs text-muted-foreground line-clamp-2">
                            {skill.description || ""}
                          </span>
                        </td>
                        <td className="px-4 py-2 text-right">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleInstall(skill)}
                            disabled={installing === skill.name || selectedAgents.length === 0}
                            title={selectedAgents.length === 0 ? "请先选择目标智能体" : "安装到选中的智能体"}
                          >
                            {installing === skill.name ? (
                              <><Loader2 className="h-3 w-3 animate-spin" /> 安装中...</>
                            ) : (
                              <><Download className="h-3 w-3" /> 安装</>
                            )}
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        );
      })}

      {/* Install Log */}
      {installLog.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">安装日志</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="bg-muted p-3 rounded-md text-xs whitespace-pre-wrap font-mono max-h-40 overflow-y-auto">
              {installLog.join("\n")}
            </pre>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
