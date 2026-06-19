import { Puzzle, Users, Package, Database, Search, FolderOpen } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { useI18n } from "../i18n/context";
import type { ListedSkill, AgentInfo, AgentGroup } from "../types";

interface Props {
  skills: ListedSkill[];
  agents: AgentInfo[];
  onNavigate: (skill: ListedSkill) => void;
  onNavigatePage?: (page: string) => void;
}

function groupAgentsByPath(agents: AgentInfo[]): AgentGroup[] {
  const groups = new Map<string, AgentGroup>();
  const detected = agents.filter(a => a.detected);
  for (const a of detected) {
    const key = a.path;
    if (!groups.has(key)) {
      groups.set(key, {
        path: key,
        agents: [{ id: a.id, name: a.name }],
        detected: true,
        displayName: a.name,
        tooltipName: a.name,
      });
    } else {
      const group = groups.get(key)!;
      group.agents.push({ id: a.id, name: a.name });
    }
  }
  for (const group of groups.values()) {
    const sorted = [...group.agents].sort((a, b) => a.name.length - b.name.length);
    group.displayName = sorted[0]?.name ?? group.agents[0]?.name ?? "";
    group.tooltipName = group.agents.map(a => a.name).join(", ");
  }
  return Array.from(groups.values());
}

export default function DashboardPage({ skills, agents, onNavigate, onNavigatePage }: Props) {
  const { t } = useI18n();
  const detectedCount = agents.filter(a => a.detected).length;
  const inPoolCount = skills.filter(s => s.inPool).length;

  const statCards = [
    { label: "技能总数", value: skills.length, icon: Puzzle, color: "text-blue-600 bg-blue-100" },
    { label: "已入池", value: inPoolCount, icon: Database, color: "text-green-600 bg-green-100" },
    { label: "智能体工具", value: detectedCount, icon: Package, color: "text-orange-600 bg-orange-100" },
    { label: "已检测智能体", value: detectedCount, icon: Users, color: "text-purple-600 bg-purple-100" },
  ];

  const groups = groupAgentsByPath(agents);

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">{t("dashboard.title")}</h2>
        <p className="text-muted-foreground mt-1">{t("dashboard.subtitle")}</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-4 gap-4">
        {statCards.map((s, i) => (
          <Card key={s.label} className={i === 0 ? "border-l-4 border-l-blue-500" : ""}>
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-sm font-medium">{s.label}</CardTitle>
              <div className={`p-2 rounded-md ${s.color}`}>
                <s.icon className="h-4 w-4" />
              </div>
            </CardHeader>
            <CardContent>
              <div className={i === 0 ? "text-3xl font-bold" : "text-2xl font-bold"}>{s.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Recent Skills */}
      <Card>
        <CardHeader>
          <CardTitle>技能概览</CardTitle>
        </CardHeader>
        <CardContent>
          {skills.length === 0 ? (
            <div className="py-8 text-center space-y-3">
              <p className="text-muted-foreground text-sm">暂无技能数据</p>
              <div className="flex justify-center gap-3">
                {onNavigatePage && (
                  <>
                    <Button variant="outline" size="sm" onClick={() => onNavigatePage("market")}>
                      <Search className="h-4 w-4 mr-1" />
                      去市场搜索
                    </Button>
                    <Button variant="outline" size="sm" onClick={() => onNavigatePage("pool")}>
                      <FolderOpen className="h-4 w-4 mr-1" />
                      扫描本地技能
                    </Button>
                  </>
                )}
              </div>
            </div>
          ) : (
            <div className="border rounded-lg overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-muted/50 border-b">
                    <th className="text-left font-medium px-4 py-2">技能名称</th>
                    <th className="text-left font-medium px-4 py-2">智能体工具</th>
                    <th className="text-left font-medium px-4 py-2">状态</th>
                    <th className="text-right font-medium px-4 py-2"></th>
                  </tr>
                </thead>
                <tbody>
                  {skills.slice(0, 20).map((s) => (
                    <tr key={s.name} className="border-b last:border-0 hover:bg-muted/30">
                      <td className="px-4 py-2 font-medium">{s.name}</td>
                      <td className="px-4 py-2">
                        <div className="flex flex-wrap gap-1">
                          {s.agentNames.map((name, i) => (
                            <Badge key={i} variant="outline" className="text-[10px]">{name}</Badge>
                          ))}
                        </div>
                      </td>
                      <td className="px-4 py-2">
                        {s.inPool ? <Badge variant="default" className="text-[10px]">池</Badge> : null}
                      </td>
                      <td className="px-4 py-2 text-right">
                        <Button variant="ghost" size="sm" onClick={() => onNavigate(s)}>详情</Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Agents Summary */}
      <Card>
        <CardHeader>
          <CardTitle>已检测智能体</CardTitle>
        </CardHeader>
        <CardContent>
          {groups.length === 0 ? (
            <p className="text-muted-foreground text-sm py-4 text-center">未检测到已安装的智能体</p>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {groups.map((group) => (
                <div key={group.path} className="border rounded-lg p-3 space-y-2 border-green-200 bg-green-50/30">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 min-w-0">
                      <span className="w-2 h-2 rounded-full shrink-0 bg-green-500" />
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
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
