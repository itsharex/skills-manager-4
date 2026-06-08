import { useState, useMemo } from "react";
import type { Agent } from "../types";

const PAGE_SIZE = 10;

interface Props {
  agents: Record<string, Agent>;
  onRefresh: () => void;
}

export default function AgentsPage({ agents, onRefresh }: Props) {
  const [search, setSearch] = useState("");
  const [visibleCount, setVisibleCount] = useState(PAGE_SIZE);

  const allIds = useMemo(() => Object.keys(agents).sort(), [agents]);

  const filteredIds = useMemo(() => {
    if (!search.trim()) return allIds;
    const q = search.toLowerCase();
    return allIds.filter((id) => {
      const a = agents[id];
      return (
        a.name.toLowerCase().includes(q) ||
        id.toLowerCase().includes(q)
      );
    });
  }, [allIds, agents, search]);

  const visibleIds = filteredIds.slice(0, visibleCount);
  const hasMore = visibleCount < filteredIds.length;

  return (
    <div>
      <header className="page-header">
        <h1 className="page-title">Agent 配置</h1>
        <p className="page-desc">
          管理支持的 Agent 列表及对应的技能目录。绿色圆点表示当前系统中已检测到。
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
              setVisibleCount(PAGE_SIZE);
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

      {filteredIds.length === 0 ? (
        <div className="card empty">
          <div className="empty-icon">🤖</div>
          <div>{search ? "未找到匹配的 Agent。" : "暂无可配置的 Agent。"}</div>
        </div>
      ) : (
        <>
          <div className="agent-grid">
            {visibleIds.map((id) => {
              const a = agents[id];
              return (
                <div key={id} className="agent-card">
                  <div className="agent-card-header">
                    <div className="agent-name">{a.name}</div>
                    <span
                      className={"dot" + (a.detected ? " on" : "")}
                      title={a.detected ? "已检测" : "未检测"}
                    />
                  </div>
                  <div className="agent-path">
                    <div>全局: {a.global_location}</div>
                    <div style={{ marginTop: 4 }}>项目: {a.skill_location}</div>
                  </div>
                  <div>
                    <span className="badge">ID: {id}</span>
                    <span
                      className="badge"
                      style={{ marginLeft: 6, background: a.detected ? "var(--success)" : "var(--bg-3)" }}
                    >
                      {a.detected ? "已检测" : "未检测"}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
          {hasMore && (
            <div style={{ textAlign: "center", marginTop: 20 }}>
              <button
                className="btn btn-secondary"
                onClick={() => setVisibleCount((c) => c + PAGE_SIZE)}
              >
                更多 ({filteredIds.length - visibleCount} 剩余)
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
