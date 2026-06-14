import { useState, useEffect } from "react";
import { Search, Loader2, AlertCircle, Globe, Database } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card";
import { Button } from "../components/ui/button";
import { Badge } from "../components/ui/badge";
import AgentSelector from "../components/AgentSelector";
import { useI18n } from "../i18n/context";
import { getConfig, searchSkills, installSkill, listPool } from "../bridge";
import type { MarketSource, ResolvedSkill, InstallResult } from "../types";

interface Props {
  onRefresh: () => void;
}

interface SearchResult {
  sourceName: string;
  sourceType: string;
  skills: ResolvedSkill[];
  error?: string;
}

const SOURCE_TYPE_LABELS: Record<string, string> = {
  pool: "本地池",
  github: "GitHub",
  registry: "开放市场",
};

const SOURCE_TYPE_ICONS: Record<string, React.ReactNode> = {
  pool: <Database className="h-4 w-4" />,
  github: <Globe className="h-4 w-4" />,
  registry: <Globe className="h-4 w-4" />,
};

export default function MarketPage({ onRefresh }: Props) {
  const { t } = useI18n();
  const [marketSources, setMarketSources] = useState<MarketSource[]>([]);
  const [loadingSources, setLoadingSources] = useState(true);
  const [searching, setSearching] = useState<string | null>(null); // source name being searched, or "all"
  const [results, setResults] = useState<SearchResult[]>([]);
  const [searchError, setSearchError] = useState<string | null>(null);
  const [selectedAgents, setSelectedAgents] = useState<string[]>([]);
  const [installing, setInstalling] = useState(false);
  const [installLog, setInstallLog] = useState<string>("");

  // Load market sources from config on mount
  useEffect(() => {
    async function load() {
      try {
        const cfg = await getConfig();
        setMarketSources(cfg.market_sources || []);
      } catch {
        // backend not available
      } finally {
        setLoadingSources(false);
      }
    }
    load();
  }, []);

  const searchSingleSource = async (source: MarketSource): Promise<SearchResult> => {
    try {
      if (source.type === "pool") {
        // Pool source: list pool and return as resolved skills
        const poolSkills = await listPool();
        const skills: ResolvedSkill[] = poolSkills.map(s => ({
          name: s.name,
          namespace: "pool",
          version: s.version || "latest",
          localPath: s.path,
        }));
        return { sourceName: source.name, sourceType: source.type, skills };
      } else {
        // GitHub/Registry: use existing search API
        const skills = await searchSkills(source.url);
        return { sourceName: source.name, sourceType: source.type, skills };
      }
    } catch (e: any) {
      return {
        sourceName: source.name,
        sourceType: source.type,
        skills: [],
        error: e.message || "搜索失败",
      };
    }
  };

  const handleSearchSource = async (source: MarketSource) => {
    setSearching(source.name);
    setSearchError(null);
    setInstallLog("");
    try {
      const result = await searchSingleSource(source);
      // Replace results for this source, keep others
      setResults(prev => {
        const filtered = prev.filter(r => r.sourceName !== source.name);
        return [...filtered, result];
      });
    } finally {
      setSearching(null);
    }
  };

  const handleSearchAll = async () => {
    const enabledSources = marketSources.filter(s => s.enabled);
    if (enabledSources.length === 0) {
      setSearchError("没有已启用的市场来源。请在设置中添加并启用来源。");
      return;
    }

    setSearching("all");
    setSearchError(null);
    setInstallLog("");
    try {
      // Search pool sources first, then others
      const ordered = [...enabledSources].sort((a, b) =>
        a.type === "pool" ? -1 : b.type === "pool" ? 1 : 0
      );

      const allResults = await Promise.all(ordered.map(s => searchSingleSource(s)));
      setResults(allResults);
    } finally {
      setSearching(null);
    }
  };

  const handleInstall = async () => {
    // Collect all skills from all results
    const allSkills = results.flatMap(r => r.skills);
    if (allSkills.length === 0) return;

    setInstalling(true);
    setInstallLog("");
    try {
      // Install from the first source that has skills
      const firstSource = results.find(r => r.skills.length > 0);
      if (!firstSource) return;

      const source = marketSources.find(s => s.name === firstSource.sourceName);
      if (!source) return;

      const res: InstallResult[] = await installSkill(source.url, { agents: selectedAgents });
      setInstallLog(res.map(r => `✅ ${r.name}@${r.version} → ${r.synced ? r.storePath : r.error}`).join("\n"));
      onRefresh();
    } catch (e: any) {
      setInstallLog(`${t("market.error_prefix")}: ${e.message}`);
    } finally {
      setInstalling(false);
    }
  };

  if (loadingSources) {
    return (
      <div className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">{t("market.title")}</h2>
          <p className="text-muted-foreground mt-1">{t("market.subtitle")}</p>
        </div>
        <Card>
          <CardContent className="py-6 text-center text-muted-foreground text-sm">
            加载市场来源...
          </CardContent>
        </Card>
      </div>
    );
  }

  const enabledSources = marketSources.filter(s => s.enabled);
  const hasResults = results.length > 0 && results.some(r => r.skills.length > 0);

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">{t("market.title")}</h2>
        <p className="text-muted-foreground mt-1">{t("market.subtitle")}</p>
      </div>

      {/* No sources configured */}
      {marketSources.length === 0 ? (
        <Card>
          <CardContent className="py-6 text-center">
            <Globe className="h-8 w-8 mx-auto text-muted-foreground mb-2" />
            <p className="text-sm text-muted-foreground mb-4">
              未配置市场来源。请在设置中添加市场来源（GitHub 仓库、本地池或开放市场）。
            </p>
          </CardContent>
        </Card>
      ) : (
        <>
          {/* Source cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {marketSources.map((source, i) => (
              <Card key={i} className={!source.enabled ? "opacity-60" : ""}>
                <CardHeader className="pb-2">
                  <CardTitle className="flex items-center gap-2 text-sm">
                    {SOURCE_TYPE_ICONS[source.type] || <Globe className="h-4 w-4" />}
                    <span className="truncate">{source.name}</span>
                    <Badge variant="outline" className="text-[10px] ml-auto">
                      {SOURCE_TYPE_LABELS[source.type] || source.type}
                    </Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <p className="text-xs text-muted-foreground truncate font-mono" title={source.url}>
                    {source.url}
                  </p>
                  {!source.enabled && (
                    <Badge variant="secondary" className="text-[10px]">已禁用</Badge>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full"
                    onClick={() => handleSearchSource(source)}
                    disabled={searching !== null || !source.enabled}
                  >
                    {searching === source.name ? (
                      <Loader2 className="h-3 w-3 animate-spin mr-1" />
                    ) : (
                      <Search className="h-3 w-3 mr-1" />
                    )}
                    搜索
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Search All Button */}
          {enabledSources.length > 1 && (
            <div className="flex justify-center">
              <Button onClick={handleSearchAll} disabled={searching !== null} variant="default">
                {searching === "all" ? (
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                ) : (
                  <Search className="h-4 w-4 mr-2" />
                )}
                搜索所有来源 ({enabledSources.length})
              </Button>
            </div>
          )}
        </>
      )}

      {/* Search error */}
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

      {/* Results grouped by source */}
      {hasResults && (
        <>
          {/* Agent selector + install button */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>{t("market.results")}</CardTitle>
              <Button onClick={handleInstall} disabled={installing}>
                {installing && <Loader2 className="h-4 w-4 animate-spin" />}
                {t("market.install_all")}
                {selectedAgents.length > 0 && (
                  <span className="ml-1 text-xs opacity-70">({selectedAgents.length})</span>
                )}
              </Button>
            </CardHeader>
            <CardContent className="space-y-4">
              <AgentSelector onSelectionChange={setSelectedAgents} />

              {/* Results by source */}
              {results.filter(r => r.skills.length > 0).map((result, ri) => (
                <div key={ri} className="space-y-2">
                  <div className="flex items-center gap-2">
                    {SOURCE_TYPE_ICONS[result.sourceType] || <Globe className="h-4 w-4" />}
                    <span className="text-sm font-medium">{result.sourceName}</span>
                    <Badge variant="outline" className="text-[10px]">{result.skills.length} 个技能</Badge>
                  </div>
                  <div className="border rounded-lg overflow-hidden">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="bg-muted/50 border-b">
                          <th className="text-left font-medium px-4 py-2">{t("market.table.name")}</th>
                          <th className="text-left font-medium px-4 py-2">{t("market.table.namespace")}</th>
                          <th className="text-left font-medium px-4 py-2">{t("market.table.version")}</th>
                          <th className="text-left font-medium px-4 py-2">{t("market.table.path")}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {result.skills.map((r, i) => (
                          <tr key={i} className="border-b last:border-0 hover:bg-muted/30">
                            <td className="px-4 py-2 font-medium">{r.name}</td>
                            <td className="px-4 py-2 text-muted-foreground">{r.namespace}</td>
                            <td className="px-4 py-2"><Badge variant="secondary">{r.version}</Badge></td>
                            <td className="px-4 py-2 text-muted-foreground font-mono text-xs">{r.localPath}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {result.error && (
                    <p className="text-xs text-destructive flex items-center gap-1">
                      <AlertCircle className="h-3 w-3" />
                      {result.error}
                    </p>
                  )}
                </div>
              ))}
            </CardContent>
          </Card>
        </>
      )}

      {/* Install log */}
      {installLog && (
        <Card>
          <CardHeader>
            <CardTitle>{t("market.install_log")}</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="bg-muted p-3 rounded-md text-xs whitespace-pre-wrap font-mono">{installLog}</pre>
          </CardContent>
        </Card>
      )}
    </div>
  );
}