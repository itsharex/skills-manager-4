import { useEffect, useState } from "react";
import type { Skill, Agent, Config } from "./types";
import InstallPage from "./pages/InstallPage";
import SkillsPage from "./pages/SkillsPage";
import AgentsPage from "./pages/AgentsPage";
import { getConfig, listSkills, listAgents } from "./bridge";

type PageKey = "install" | "skills" | "agents";

interface NavItem {
  key: PageKey;
  label: string;
  icon: string;
}

const NAV: NavItem[] = [
  { key: "install", label: "安装技能", icon: "⬇" },
  { key: "skills", label: "已安装技能", icon: "📦" },
  { key: "agents", label: "Agent 配置", icon: "🤖" },
];

export default function App() {
  const [page, setPage] = useState<PageKey>("install");
  const [skills, setSkills] = useState<Skill[]>([]);
  const [agents, setAgents] = useState<Record<string, Agent>>({});
  const [config, setConfig] = useState<Config | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    (async () => {
      const [cfg, sk, ag] = await Promise.all([
        getConfig(),
        listSkills(),
        listAgents(),
      ]);
      setConfig(cfg);
      setSkills(sk);
      setAgents(ag);
    })();
  }, [refreshKey]);

  const refresh = () => setRefreshKey((n) => n + 1);

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-title">
            <div className="brand-icon">S</div>
            <span>Skills Manager</span>
          </div>
          <div className="brand-sub">v0.2 · 跨 Agent 技能管理</div>
        </div>
        <nav className="nav">
          {NAV.map((n) => (
            <div
              key={n.key}
              className={"nav-item" + (page === n.key ? " active" : "")}
              onClick={() => setPage(n.key)}
            >
              <div className="nav-icon">{n.icon}</div>
              <span>{n.label}</span>
            </div>
          ))}
        </nav>
        <div className="sidebar-footer">
          <div className="skillspool-info">
            📁 skillspool
            <br />
            {config?.skillspool?.root || "未配置"}
          </div>
        </div>
      </aside>

      <main className="main">
        <div className="page">
          {page === "install" && (
            <InstallPage
              agents={agents}
              onInstalled={refresh}
            />
          )}
          {page === "skills" && <SkillsPage skills={skills} onRefresh={refresh} />}
          {page === "agents" && <AgentsPage agents={agents} onRefresh={refresh} />}
        </div>
      </main>
    </div>
  );
}
