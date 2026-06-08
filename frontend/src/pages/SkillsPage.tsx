import type { Skill } from "../types";

interface Props {
  skills: Skill[];
  onRefresh: () => void;
}

export default function SkillsPage({ skills, onRefresh }: Props) {
  return (
    <div>
      <header className="page-header">
        <h1 className="page-title">已安装技能</h1>
        <p className="page-desc">
          当前共 {skills.length} 个技能。点击刷新获取最新状态。
        </p>
        <button
          className="btn btn-secondary"
          onClick={onRefresh}
          style={{ marginTop: 12 }}
        >
          🔄 刷新
        </button>
      </header>

      {skills.length === 0 ? (
        <div className="card empty">
          <div className="empty-icon">📦</div>
          <div>暂无已安装的技能。前往"安装技能"添加吧。</div>
        </div>
      ) : (
        <div className="skill-grid">
          {skills.map((s) => (
            <div key={s.name} className="skill-card">
              <div className="skill-name">{s.name}</div>
              <div className="skill-desc">{s.description || "—"}</div>
              <div className="skill-meta">
                <div>
                  <span className="badge latest">latest: {s.latest_version}</span>
                  <span style={{ marginLeft: 6 }}>
                    {Object.keys(s.versions).length} 个版本
                  </span>
                </div>
              </div>
              <div style={{ marginTop: 12, fontSize: 12, color: "var(--fg-dim)" }}>
                分发到:{" "}
                {collectAgentsFromVersions(s).length > 0 ? (
                  collectAgentsFromVersions(s).map((a) => (
                    <span
                      key={a}
                      className="badge"
                      style={{ marginRight: 4, marginTop: 4, display: "inline-block" }}
                    >
                      {a}
                    </span>
                  ))
                ) : (
                  <span style={{ color: "var(--fg-dim)" }}>—</span>
                )}
              </div>
              <div style={{ marginTop: 12, fontSize: 11, color: "var(--fg-dim)", fontFamily: "ui-monospace, monospace", wordBreak: "break-all" }}>
                来源: {formatSource(s.source)}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function collectAgentsFromVersions(s: Skill): string[] {
  const set = new Set<string>();
  for (const v of Object.values(s.versions)) {
    for (const a of v.agents) {
      set.add(a);
    }
  }
  return Array.from(set).sort();
}

function formatSource(src: Skill["source"]): string {
  if (src.url) return src.url;
  if (src.path) return "local: " + src.path;
  if (src.command) return src.command;
  return "-";
}
