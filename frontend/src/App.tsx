import { useEffect, useState } from "react";
import type { Agent, Config } from "./types";
import InstallPage from "./pages/InstallPage";
import SkillsPage from "./pages/SkillsPage";
import AgentsPage from "./pages/AgentsPage";
import ClawHubPage from "./pages/ClawHubPage";
import { getConfig, listAgents, getSkillspoolRoot, setSkillspoolRoot, selectDirectory } from "./bridge";

type PageKey = "skills" | "install" | "agents" | "clawhub";

interface NavItem {
  key: PageKey;
  label: string;
  icon: string;
}

const NAV: NavItem[] = [
  { key: "skills",   label: "技能库",       icon: "📦" },  // 技能库放第一位
  { key: "install",  label: "安装技能",     icon: "⬇" },
  { key: "agents",   label: "Agent 配置",   icon: "🤖" },
  { key: "clawhub",  label: "ClawHub 市场", icon: "🦞" },  // ClawHub市场放最后
];

export default function App() {
  const [page, setPage] = useState<PageKey>("skills");  // 默认显示技能库
  const [agents, setAgents] = useState<Record<string, Agent>>({});
  const [config, setConfig] = useState<Config | null>(null);
  const [skillspoolRoot, setSkillspoolRootState] = useState<string>("");
  const [refreshKey, setRefreshKey] = useState(0);
  // 迁移/编辑技能池路径弹窗状态
  const [showEditPool, setShowEditPool] = useState(false);
  const [editPoolPath, setEditPoolPath] = useState<string>("");
  const [editPoolBusy, setEditPoolBusy] = useState(false);
  const [editPoolErr, setEditPoolErr] = useState<string | null>(null);
  const [editPoolInfo, setEditPoolInfo] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      const [cfg, ag, root] = await Promise.all([getConfig(), listAgents(), getSkillspoolRoot()]);
      setConfig(cfg);
      setAgents(ag);
      setSkillspoolRootState(root || cfg?.skillspool?.root || "");
    })();
  }, [refreshKey]);

  const refresh = () => setRefreshKey((n) => n + 1);

  const openEditPool = () => {
    setEditPoolPath(skillspoolRoot || "");
    setEditPoolErr(null);
    setEditPoolInfo(null);
    setShowEditPool(true);
  };

  const pickPoolPath = async () => {
    try {
      const result = await selectDirectory("选择新的技能池根目录");
      if (result) setEditPoolPath(result);
    } catch (e: any) {
      setEditPoolErr(String(e?.message || e));
    }
  };

  const submitEditPool = async () => {
    if (!editPoolPath.trim()) {
      setEditPoolErr("路径不能为空");
      return;
    }
    if (editPoolPath === skillspoolRoot) {
      setShowEditPool(false);
      return;
    }
    setEditPoolBusy(true);
    setEditPoolErr(null);
    setEditPoolInfo(null);
    try {
      const result: any = await setSkillspoolRoot(editPoolPath.trim());
      if (result && result.success) {
        setSkillspoolRootState(result.new_root || editPoolPath);
        setEditPoolInfo(result.message || "迁移成功");
        // 自动刷新
        setTimeout(() => {
          refresh();
        }, 1500);
      } else {
        setEditPoolErr(result?.message || "迁移失败");
      }
    } catch (e: any) {
      setEditPoolErr(String(e?.message || e));
    } finally {
      setEditPoolBusy(false);
    }
  };

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
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 4 }}>
              <span>📁 skillspool</span>
              <button
                className="icon-btn"
                onClick={openEditPool}
                title="修改技能池路径"
                style={{ background: "transparent", border: "none", color: "var(--accent)", cursor: "pointer", fontSize: 12, padding: "2px 6px" }}
              >
                ✏️ 修改
              </button>
            </div>
            <div style={{ fontSize: 11, wordBreak: "break-all", opacity: 0.85 }}>
              {skillspoolRoot || config?.skillspool?.root || "未配置"}
            </div>
          </div>
        </div>
      </aside>

      <main className="main">
        <div className="page">
          {page === "install" && (
            <InstallPage
              onInstalled={refresh}
            />
          )}
          {page === "skills" && <SkillsPage agents={agents} onRefresh={refresh} />}
          {page === "clawhub" && <ClawHubPage />}
          {page === "agents" && <AgentsPage agents={agents} onRefresh={refresh} />}
        </div>
      </main>

      {/* 修改技能池路径弹窗 */}
      {showEditPool && (
        <div
          style={{
            position: "fixed", inset: 0, background: "rgba(0,0,0,0.6)",
            display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1000,
          }}
          onClick={() => !editPoolBusy && setShowEditPool(false)}
        >
          <div
            className="card"
            style={{ width: 540, maxWidth: "92vw", padding: 24 }}
            onClick={(e) => e.stopPropagation()}
          >
            <h2 style={{ marginTop: 0, fontSize: 18 }}>📁 修改技能池根目录</h2>
            <p style={{ color: "var(--fg-dim)", fontSize: 13, margin: "4px 0 16px" }}>
              当前路径：<code style={{ background: "var(--bg)", padding: "2px 6px", borderRadius: 4 }}>{skillspoolRoot || "未配置"}</code>
            </p>
            <p style={{ color: "var(--accent)", fontSize: 12, margin: "0 0 12px", lineHeight: 1.6 }}>
              ⚠️ 修改路径将自动迁移所有技能（先复制到新位置，验证后删除旧位置）。<br />
              迁移过程中请勿关闭应用。失败时原始数据不会被破坏。
            </p>

            <div className="form-group">
              <label className="form-label">新路径</label>
              <div style={{ display: "flex", gap: 8 }}>
                <input
                  className="form-input"
                  value={editPoolPath}
                  onChange={(e) => setEditPoolPath(e.target.value)}
                  placeholder="例如: D:\MySkills\skillspool"
                  disabled={editPoolBusy}
                  spellCheck={false}
                />
                <button
                  className="btn btn-secondary"
                  onClick={pickPoolPath}
                  disabled={editPoolBusy}
                  type="button"
                >
                  📁 浏览
                </button>
              </div>
            </div>

            {editPoolErr && (
              <div
                style={{
                  padding: 10, marginTop: 8, borderRadius: 6,
                  background: "rgba(239, 68, 68, 0.12)",
                  color: "var(--danger)", fontSize: 13,
                }}
              >
                ❌ {editPoolErr}
              </div>
            )}
            {editPoolInfo && (
              <div
                style={{
                  padding: 10, marginTop: 8, borderRadius: 6,
                  background: "rgba(34, 197, 94, 0.12)",
                  color: "var(--success)", fontSize: 13,
                }}
              >
                ✅ {editPoolInfo}
              </div>
            )}

            <div style={{ display: "flex", gap: 8, marginTop: 18, justifyContent: "flex-end" }}>
              <button
                className="btn btn-secondary"
                onClick={() => setShowEditPool(false)}
                disabled={editPoolBusy}
              >
                取消
              </button>
              <button
                className="btn btn-primary"
                onClick={submitEditPool}
                disabled={editPoolBusy}
              >
                {editPoolBusy ? "⏳ 迁移中..." : "开始迁移"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
