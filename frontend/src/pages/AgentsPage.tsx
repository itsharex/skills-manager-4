import { useState, useMemo } from "react";
import type { Agent } from "../types";

const PAGE_SIZE = 10;

interface Props {
  agents: Record<string, Agent>;
  onRefresh: () => void;
}

export default function AgentsPage({ agents, onRefresh }: Props) {
  const [search, setSearch] = useState("");
  const [installedVisible, setInstalledVisible] = useState(PAGE_SIZE);
  const [notInstalledVisible, setNotInstalledVisible] = useState(PAGE_SIZE);

  // 将 Agent 分为已安装和未安装两组
  const { installedIds, notInstalledIds } = useMemo(() => {
    const installed: string[] = [];
    const notInstalled: string[] = [];
    Object.keys(agents).forEach((id) => {
      const a = agents[id];
      // 已安装：detected=true 且 installed=true（目录存在且有技能）
      // 或者 detected=true（至少目录存在）
      if (a.detected || a.installed) {
        installed.push(id);
      } else {
        notInstalled.push(id);
      }
    });
    // 已安装组：按名称字母序
    installed.sort((a, b) => agents[a].name.localeCompare(agents[b].name));
    // 未安装组：按名称字母序
    notInstalled.sort((a, b) => agents[a].name.localeCompare(agents[b].name));
    return { installedIds: installed, notInstalledIds: notInstalled };
  }, [agents]);

  // 搜索过滤（跨两组）
  const filteredInstalled = useMemo(() => {
    if (!search.trim()) return installedIds;
    const q = search.toLowerCase();
    return installedIds.filter((id) => {
      const a = agents[id];
      return a.name.toLowerCase().includes(q) || id.toLowerCase().includes(q);
    });
  }, [installedIds, agents, search]);

  const filteredNotInstalled = useMemo(() => {
    if (!search.trim()) return notInstalledIds;
    const q = search.toLowerCase();
    return notInstalledIds.filter((id) => {
      const a = agents[id];
      return a.name.toLowerCase().includes(q) || id.toLowerCase().includes(q);
    });
  }, [notInstalledIds, agents, search]);

  const installedVisibleIds = filteredInstalled.slice(0, installedVisible);
  const installedHasMore = installedVisible < filteredInstalled.length;

  const notInstalledVisibleIds = filteredNotInstalled.slice(0, notInstalledVisible);
  const notInstalledHasMore = notInstalledVisible < filteredNotInstalled.length;

  // 汇总
  const totalAgents = Object.keys(agents).length;

  return (
    <div>
      <header className="page-header">
        <h1 className="page-title">Agent 配置</h1>
        <p className="page-desc">
          管理支持的 Agent 列表及对应的技能目录。
          当前共 <strong>{totalAgents}</strong> 个 Agent，
          本地已安装 <strong style={{ color: "var(--success)" }}>{installedIds.length}</strong>，
          未安装 <strong style={{ color: "var(--fg-dim)" }}>{notInstalledIds.length}</strong>。
        </p>
        <div className="search-bar">
          <span className="search-icon">🔍</span>
          <input
            type="text"
            className="search-input"
            placeholder="搜索 Agent..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setInstalledVisible(PAGE_SIZE);
              setNotInstalledVisible(PAGE_SIZE);
            }}
          />
        </div>
        <button
          className="btn btn-secondary"
          onClick={onRefresh}
          style={{ marginTop: 12 }}
        >
          🔄 重新检测
        </button>
      </header>

      {totalAgents === 0 ? (
        <div className="card empty">
          <div className="empty-icon">🤖</div>
          <div>{search ? "未找到匹配的 Agent。" : "暂无可配置的 Agent。"}</div>
        </div>
      ) : (
        <>
          {/* ============ 已安装 Agent ============ */}
          <section style={{ marginBottom: 32 }}>
            <div className="agent-section-header">
              <h2 className="agent-section-title">
                <span style={{ color: "var(--success)" }}>●</span>
                本地已安装
                <span className="agent-section-count">({filteredInstalled.length})</span>
              </h2>
              <p className="agent-section-desc">
                已检测到目录存在或已安装过技能的 Agent（按名称排序）
              </p>
            </div>

            {filteredInstalled.length === 0 ? (
              <div className="card empty" style={{ padding: "24px" }}>
                <div className="empty-icon">✅</div>
                <div>所有已安装 Agent 已显示完毕。</div>
              </div>
            ) : (
              <>
                <div className="agent-grid">
                  {installedVisibleIds.map((id) => (
                    <AgentCard key={id} id={id} agent={agents[id]} />
                  ))}
                </div>
                {installedHasMore && (
                  <div style={{ textAlign: "center", marginTop: 16 }}>
                    <button
                      className="btn btn-secondary"
                      onClick={() => setInstalledVisible((c) => c + PAGE_SIZE)}
                    >
                      更多已安装 Agent ({filteredInstalled.length - installedVisibleIds.length} 剩余)
                    </button>
                  </div>
                )}
              </>
            )}
          </section>

          {/* ============ 未安装 Agent ============ */}
          <section>
            <div className="agent-section-header">
              <h2 className="agent-section-title">
                <span style={{ color: "var(--fg-dim)" }}>○</span>
                未安装
                <span className="agent-section-count">({filteredNotInstalled.length})</span>
              </h2>
              <p className="agent-section-desc">
                未检测到目录或未安装过技能的 Agent
              </p>
            </div>

            {filteredNotInstalled.length === 0 ? (
              <div className="card empty" style={{ padding: "24px" }}>
                <div className="empty-icon">🎉</div>
                <div>{search ? "搜索结果为空" : "暂无未安装的 Agent"}</div>
              </div>
            ) : (
              <>
                <div className="agent-grid">
                  {notInstalledVisibleIds.map((id) => (
                    <AgentCard key={id} id={id} agent={agents[id]} />
                  ))}
                </div>
                {notInstalledHasMore && (
                  <div style={{ textAlign: "center", marginTop: 16 }}>
                    <button
                      className="btn btn-secondary"
                      onClick={() => setNotInstalledVisible((c) => c + PAGE_SIZE)}
                    >
                      更多未安装 Agent ({filteredNotInstalled.length - notInstalledVisibleIds.length} 剩余)
                    </button>
                  </div>
                )}
              </>
            )}
          </section>
        </>
      )}

      <style>{styles}</style>
    </div>
  );
}

// Agent 卡片组件
function AgentCard({ id, agent }: { id: string; agent: Agent }) {
  return (
    <div className={"agent-card" + (agent.detected ? " detected" : "")}>
      <div className="agent-card-header">
        <div className="agent-name">
          {agent.name}
          {agent.installed && (
            <span
              className="status-mini"
              style={{ background: "var(--primary)", color: "#fff" }}
            >
              已安装
            </span>
          )}
          {agent.detected && (
            <span
              className="status-mini"
              style={{ background: "var(--success)", color: "#fff" }}
            >
              已检测
            </span>
          )}
          {!agent.detected && !agent.installed && (
            <span
              className="status-mini"
              style={{ background: "var(--bg-3)", color: "var(--fg-dim)" }}
            >
              未检测
            </span>
          )}
        </div>
        <span
          className={"dot" + (agent.detected ? " on" : "")}
          title={agent.detected ? "已检测" : "未检测"}
        />
      </div>
      <div className="agent-path">
        <div>全局: <code>{agent.global_location || "—"}</code></div>
        <div style={{ marginTop: 4 }}>项目: <code>{agent.skill_location || "—"}</code></div>
      </div>
      <div>
        <span className="badge">ID: {id}</span>
        {!agent.supports_project && (
          <span
            className="badge"
            style={{ marginLeft: 6, background: "rgba(255, 180, 0, 0.2)", color: "#f59e0b" }}
            title="该 Agent 仅支持全局安装"
          >
            仅全局
          </span>
        )}
      </div>
    </div>
  );
}

const styles = `
.agent-section-header {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border);
}
.agent-section-title {
  font-size: 16px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
}
.agent-section-count {
  font-size: 13px;
  font-weight: 400;
  color: var(--fg-dim);
}
.agent-section-desc {
  font-size: 13px;
  color: var(--fg-dim);
  margin: 4px 0 0 0;
}

.agent-card {
  padding: 14px;
  background: var(--bg-2);
  border: 1px solid var(--border);
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.agent-card.detected {
  border-color: var(--success);
  box-shadow: 0 0 0 1px rgba(67, 202, 137, 0.15);
}
.agent-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.agent-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  font-size: 14px;
}
.status-mini {
  font-size: 10px;
  padding: 2px 7px;
  border-radius: 999px;
  font-weight: 500;
}
.agent-path {
  font-size: 12px;
  color: var(--fg-dim);
  background: var(--bg);
  padding: 8px 10px;
  border-radius: 6px;
  border: 1px solid var(--border);
  line-height: 1.5;
}
.agent-path code {
  font-family: ui-monospace, monospace;
  color: var(--fg);
}
.agent-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 12px;
}
`;
