import { ArrowLeft, Trash2, MapPin, ExternalLink, Loader2, FolderOpen } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "../components/ui/tabs";
import { useI18n } from "../i18n/context";
import { deleteSkill, openDirectory, getConfig } from "../bridge";
import type { ListedSkill } from "../types";
import { useState, useEffect } from "react";

interface Props {
  skill: ListedSkill | null;
  onBack: () => void;
}

export default function DetailPage({ skill, onBack }: Props) {
  const { t } = useI18n();
  const [deleting, setDeleting] = useState<string | null>(null);
  const [deleteMsg, setDeleteMsg] = useState<string | null>(null);
  const [openingFolder, setOpeningFolder] = useState<string | null>(null);
  const [poolPath, setPoolPath] = useState("");

  useEffect(() => {
    getConfig().then(cfg => {
      if (cfg?.pool_path) setPoolPath(cfg.pool_path);
    }).catch(() => {});
  }, []);

  if (!skill) {
    return (
      <div className="space-y-6">
        <h2 className="text-2xl font-bold tracking-tight">{t("detail.title")}</h2>
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            {t("detail.no_selection")}
          </CardContent>
        </Card>
      </div>
    );
  }

  const versions = skill.versions ?? [];

  const handleDelete = async (path: string, agentName: string) => {
    if (!confirm(`确认删除 ${agentName} 下的技能 "${skill.name}"？\n路径: ${path}\n此操作不可恢复。`)) return;
    setDeleting(path);
    setDeleteMsg(null);
    try {
      await deleteSkill(path);
      setDeleteMsg(`已从 ${agentName} 删除技能 "${skill.name}"`);
    } catch (e: any) {
      setDeleteMsg(`删除失败: ${e.message}`);
    } finally {
      setDeleting(null);
    }
  };

  const handleOpen = async (path: string) => {
    setOpeningFolder(path);
    try {
      await openDirectory(path);
    } catch { /* ignore */ }
    finally { setOpeningFolder(null); }
  };

  // Pool path
  const poolSkillPath = poolPath && skill.inPool ? `${poolPath}/${skill.name}` : "";
  // Agent install paths (paths indexed by agent)
  const agentPaths = skill.paths.map((path, i) => ({
    path,
    agentName: skill.agentNames[i] || skill.agentIds[i] || "未知",
  }));

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={onBack}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex-1">
          <h2 className="text-2xl font-bold tracking-tight">{skill.name}</h2>
          <div className="flex items-center gap-2 mt-1">
            {skill.inPool && <Badge variant="default" className="text-[10px]">在池中</Badge>}
            {skill.agentNames.map((name, i) => (
              <Badge key={i} variant="outline" className="text-[10px]">{name}</Badge>
            ))}
          </div>
        </div>
      </div>

      {deleteMsg && (
        <Card>
          <CardContent className="py-3">
            <p className={`text-sm ${deleteMsg.includes("失败") ? "text-destructive" : "text-green-600"}`}>{deleteMsg}</p>
          </CardContent>
        </Card>
      )}

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">{t("detail.overview")}</TabsTrigger>
          <TabsTrigger value="paths">安装位置</TabsTrigger>
          <TabsTrigger value="versions">{t("detail.versions")}</TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <Card>
            <CardHeader>
              <CardTitle>{t("detail.overview")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-muted-foreground">技能名称</p>
                  <p className="font-medium">{skill.name}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">智能体工具</p>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {skill.agentNames.map((name, i) => (
                      <Badge key={i} variant="secondary">{name}</Badge>
                    ))}
                  </div>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">{t("detail.field.latest_version")}</p>
                  <Badge variant="secondary">{skill.latest || "-"}</Badge>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">{t("detail.field.total_versions")}</p>
                  <p className="font-medium">{versions.length}</p>
                </div>
              </div>
              <div>
                <p className="text-sm text-muted-foreground mb-1">{t("detail.field.description")}</p>
                <p className="text-sm">{skill.description || t("detail.no_description")}</p>
              </div>
              {poolSkillPath && (
                <div>
                  <p className="text-sm text-muted-foreground mb-1">池路径</p>
                  <div className="flex items-center gap-2">
                    <code className="text-xs bg-muted px-2 py-1 rounded font-mono flex-1 truncate">{poolSkillPath}</code>
                    <Button variant="outline" size="sm" className="h-7 text-xs shrink-0" onClick={() => handleOpen(poolSkillPath)} disabled={openingFolder === poolSkillPath}>
                      {openingFolder === poolSkillPath ? <Loader2 className="h-3 w-3 animate-spin" /> : <FolderOpen className="h-3 w-3" />}
                      打开
                    </Button>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="paths">
          <Card>
            <CardHeader>
              <CardTitle>安装位置</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {/* Pool path */}
                {poolSkillPath && (
                  <div>
                    <p className="text-xs font-medium text-muted-foreground mb-2 flex items-center gap-1">
                      <FolderOpen className="h-3 w-3" />
                      技能池位置
                    </p>
                    <div className="border rounded-lg p-3 flex items-center justify-between gap-3">
                      <div className="min-w-0 flex-1 flex items-center gap-2">
                        <MapPin className="h-3.5 w-3.5 text-green-500 shrink-0" />
                        <Badge variant="default" className="text-[10px] shrink-0">池</Badge>
                        <p className="text-xs font-mono text-muted-foreground truncate" title={poolSkillPath}>{poolSkillPath}</p>
                      </div>
                      <Button variant="outline" size="sm" className="h-7 text-xs shrink-0" onClick={() => handleOpen(poolSkillPath)} disabled={openingFolder === poolSkillPath}>
                        {openingFolder === poolSkillPath ? <Loader2 className="h-3 w-3 animate-spin" /> : <ExternalLink className="h-3 w-3" />}
                        打开
                      </Button>
                    </div>
                  </div>
                )}

                {/* Agent install paths */}
                {agentPaths.length > 0 && (
                  <div>
                    <p className="text-xs font-medium text-muted-foreground mb-2 flex items-center gap-1">
                      <MapPin className="h-3 w-3" />
                      智能体安装位置
                    </p>
                    <div className="space-y-2">
                      {agentPaths.map(({ path, agentName }, i) => (
                        <div key={i} className="border rounded-lg p-3 flex items-center justify-between gap-3">
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-2 mb-1">
                              <MapPin className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                              <Badge variant="outline" className="text-[10px]">{agentName}</Badge>
                            </div>
                            <p className="text-xs font-mono text-muted-foreground truncate" title={path}>{path}</p>
                          </div>
                          <div className="flex items-center gap-2 shrink-0">
                            <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => handleOpen(path)} disabled={openingFolder === path}>
                              {openingFolder === path ? <Loader2 className="h-3 w-3 animate-spin" /> : <ExternalLink className="h-3 w-3" />}
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-destructive hover:text-destructive hover:bg-destructive/10"
                              onClick={() => handleDelete(path, agentName)}
                              disabled={deleting === path}
                            >
                              {deleting === path ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Trash2 className="h-3.5 w-3.5" />}
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {!poolSkillPath && agentPaths.length === 0 && (
                  <p className="text-muted-foreground text-sm">无安装位置信息</p>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="versions">
          <Card>
            <CardHeader>
              <CardTitle>{t("detail.versions")}</CardTitle>
            </CardHeader>
            <CardContent>
              {versions.length === 0 ? (
                <p className="text-muted-foreground text-sm">{t("detail.no_versions")}</p>
              ) : (
                <div className="border rounded-lg overflow-hidden">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="bg-muted/50 border-b">
                        <th className="text-left font-medium px-4 py-2">{t("detail.table.version")}</th>
                        <th className="text-left font-medium px-4 py-2">{t("detail.table.status")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {versions.map((v) => (
                        <tr key={v} className="border-b last:border-0">
                          <td className="px-4 py-2 font-mono text-sm">{v}</td>
                          <td className="px-4 py-2">
                            {v === skill.latest ? <Badge>{t("detail.status_latest")}</Badge> : <Badge variant="secondary">{v}</Badge>}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}