import { useEffect, useState, useCallback } from "react";
import {
  runtimeStatus,
  ensureRuntime,
  searchClawHub,
  installFromClawHub,
  listAgents,
} from "../bridge";
import type { ClawHubSkill, RuntimeStatus, Agent, InstallResult } from "../types";

type Status = "idle" | "loading" | "success" | "error";

export default function ClawHubPage() {
  const [status, setStatus] = useState<RuntimeStatus | null>(null);
  const [keyword, setKeyword] = useState("");
  const [searching, setSearching] = useState(false);
  const [skills, setSkills] = useState<ClawHubSkill[]>([]);
  const [agents, setAgents] = useState<Record<string, Agent>>({});
  const [selectedAgents, setSelectedAgents] = useState<Record<string, boolean>>({});
  const [installStates, setInstallStates] = useState<Record<string, Status>>({});
  const [installResults, setInstallResults] = useState<Record<string, InstallResult>>({});
  const [installError, setInstallError] = useState<Record<string, string>>({});
  const [runtimeBusy, setRuntimeBusy] = useState(false);
  const [runtimeMsg, setRuntimeMsg] = useState<string>("");
  const [hasSearched, setHasSearched] = useState(false);
  const [searchError, setSearchError] = useState<string>("");

  // 初次加载
  useEffect(() => {
    runtimeStatus().then((s) => setStatus(s));
    listAgents().then((a) => {
      setAgents(a);
      const sel: Record<string, boolean> = {};
      for (const id in a) if (a[id].detected) sel[id] = true;
      setSelectedAgents(sel);
    });
  }, []);

  // 搜索
  const onSearch = useCallback(async () => {
    setSearching(true);
    setSearchError("");
    setHasSearched(true);
    try {
      const s = await searchClawHub(keyword.trim());
      setSkills(Array.isArray(s) ? s : []);
    } catch (e: any) {
      setSearchError(e?.message || String(e));
      setSkills([]);
    } finally {
      setSearching(false);
    }
  }, [keyword]);

  // 安装 clawhub CLI
  const onInstallRuntime = async () => {
    setRuntimeBusy(true);
    setRuntimeMsg("正在安装 clawhub CLI（npm install -g clawhub）…");
    try {
      const s = await ensureRuntime();
      setStatus(s);
      setRuntimeMsg(s?.message || "安装完成");
    } catch (e: any) {
      setRuntimeMsg("安装失败：" + (e?.message || String(e)));
    } finally {
      setRuntimeBusy(false);
    }
  };

  // 安装单个技能
  const onInstallSkill = async (skill: ClawHubSkill) => {
    const key = `${skill.owner}/${skill.slug}`;
    const chosen = Object.keys(selectedAgents).filter((id) => selectedAgents[id]);
    if (chosen.length === 0) {
      setInstallError((prev) => ({ ...prev, [key]: "请先勾选至少一个目标 Agent" }));
      return;
    }
    setInstallStates((p) => ({ ...p, [key]: "loading" }));
    setInstallError((p) => ({ ...p, [key]: "" }));
    try {
      const result = await installFromClawHub(skill.owner, skill.slug, chosen);
      setInstallResults((p) => ({ ...p, [key]: result }));
      setInstallStates((p) => ({ ...p, [key]: "success" }));
    } catch (e: any) {
      setInstallStates((p) => ({ ...p, [key]: "error" }));
      setInstallError((p) => ({ ...p, [key]: e?.message || String(e) }));
    }
  };

  // 热门关键词（点击即搜索）
  const hotKeywords = ["react", "python", "aws", "docker", "git", "test", "frontend", "backend"];

  const detectedAgentIds = Object.keys(agents).filter((id) => agents[id].detected);
  // 是否使用 GitHub-only 模式（无 CLI 但有 Node.js）
  const githubMode = !!(status?.nodeInstalled && !status.clawhubInstalled);
  const canSearch = !!(status?.nodeInstalled);

  return (
    <div>
      <header className="page-header">
        <h1 className="page-title">🦞 ClawHub 市场</h1>
        <p className="page-desc">
          浏览并安装 OpenClaw 社区的技能。下载的技能将以<b>目录复制</b>方式保存到全局 skillspool，
          与 GitHub/local 技能共用一套管理逻辑。
          {githubMode && (
            <span style={{ display: "block", marginTop: 4, color: "var(--accent)" }}>
              ℹ️ 当前使用 GitHub-only 模式（直接通过 GitHub API 浏览 registry，无需 CLI）。
            </span>
          )}
        </p>
      </header>

      {/* 运行时状态卡片 */}
      <section className="card" style={{ marginBottom: 20 }}>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 16 }}>
          <div style={{ flex: 1 }}>
            <div className="card-title" style={{ marginBottom: 12 }}>🛠 运行时检查</div>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 12 }}>
              <RuntimeBox label="Node.js" ok={status?.nodeInstalled} version={status?.nodeVersion} />
              <RuntimeBox label="npm" ok={status?.hasNpm} />
              <RuntimeBox label="clawhub CLI" ok={status?.clawhubInstalled} version={status?.clawhubVersion} />
            </div>
            {status?.registryReachable !== undefined && (
              <div style={{ marginTop: 8, fontSize: 12, color: "var(--fg-dim)" }}>
                Registry <code>{status.registryName || "openclaw-dev/skills"}</code>：
                {status.registryReachable ? (
                  <span style={{ color: "var(--success)" }}> ✅ 可访问</span>
                ) : (
                  <span style={{ color: "var(--danger)" }}> ❌ 不可访问</span>
                )}
              </div>
            )}
            {status?.message && (
              <div style={{ fontSize: 12, color: "var(--fg-dim)", marginTop: 6 }}>{status.message}</div>
            )}
            {runtimeMsg && (
              <div style={{ fontSize: 12, color: "var(--accent)", marginTop: 4 }}>{runtimeMsg}</div>
            )}
          </div>
          <div>
            {!status?.nodeInstalled && (
              <div style={{ fontSize: 12, color: "var(--danger)" }}>请先安装 Node.js ≥ 20 并加入 PATH</div>
            )}
            {status?.nodeInstalled && !status.clawhubInstalled && (
              <button
                className="btn btn-primary"
                onClick={onInstallRuntime}
                disabled={runtimeBusy}
              >
                {runtimeBusy ? "安装中…" : "安装 clawhub CLI"}
              </button>
            )}
            {status?.nodeInstalled && status.clawhubInstalled && (
              <div style={{ fontSize: 13, color: "var(--success)", fontWeight: 600 }}>✓ 运行时就绪</div>
            )}
          </div>
        </div>
      </section>

      {/* 搜索栏 */}
      <div className="search-bar" style={{ marginBottom: 12 }}>
        <span className="search-icon">🔍</span>
        <input
          type="text"
          className="search-input"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && canSearch && !searching) onSearch();
          }}
          placeholder="输入关键词（owner/slug/name/description）"
          disabled={!canSearch}
        />
        <button
          className="btn btn-primary"
          onClick={onSearch}
          disabled={!canSearch || searching}
        >
          {searching ? "⏳ 搜索中…" : "搜索"}
        </button>
        <button
          className="btn btn-secondary"
          onClick={() => { setKeyword(""); onSearch(); }}
          disabled={!canSearch || searching}
        >
          🔄 浏览全部
        </button>
      </div>

      {/* 热门关键词 */}
      {!hasSearched && canSearch && (
        <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginBottom: 16, alignItems: "center" }}>
          <span style={{ fontSize: 12, color: "var(--fg-dim)" }}>热门：</span>
          {hotKeywords.map((kw) => (
            <button
              key={kw}
              className="chip"
              onClick={() => { setKeyword(kw); setTimeout(onSearch, 0); }}
              type="button"
            >
              {kw}
            </button>
          ))}
        </div>
      )}

      {/* Agent 多选 */}
      <div style={{ marginBottom: 16, padding: "10px 14px", background: "var(--bg-2)", borderRadius: 8, border: "1px solid var(--border)" }}>
        <div style={{ fontSize: 12, fontWeight: 600, marginBottom: 6, color: "var(--fg-dim)" }}>
          🎯 目标 Agent（已选 {Object.values(selectedAgents).filter(Boolean).length} / {detectedAgentIds.length}）
        </div>
        {detectedAgentIds.length === 0 ? (
          <div style={{ fontSize: 13, color: "var(--fg-dim)" }}>暂未检测到任何 Agent</div>
        ) : (
          <div style={{ display: "flex", flexWrap: "wrap", gap: 10 }}>
            {detectedAgentIds.map((id) => (
              <label key={id} style={{ display: "inline-flex", alignItems: "center", gap: 4, fontSize: 13, cursor: "pointer" }}>
                <input
                  type="checkbox"
                  checked={!!selectedAgents[id]}
                  onChange={(e) =>
                    setSelectedAgents((p) => ({ ...p, [id]: e.target.checked }))
                  }
                />
                <span>{agents[id].name}</span>
              </label>
            ))}
          </div>
        )}
      </div>

      {/* 错误提示 */}
      {searchError && (
        <div className="card" style={{ background: "rgba(239, 68, 68, 0.1)", borderColor: "var(--danger)", color: "var(--danger)" }}>
          ❌ 搜索失败：{searchError}
        </div>
      )}

      {/* 技能列表 */}
      {searching ? (
        <div className="card empty" style={{ padding: 40 }}>
          <div className="empty-icon">⏳</div>
          <div>正在加载 ClawHub 技能列表...</div>
          <div style={{ fontSize: 12, color: "var(--fg-dim)", marginTop: 8 }}>
            首次加载会扫描整个 registry 仓库，请稍候
          </div>
        </div>
      ) : !hasSearched ? (
        <div className="card empty" style={{ padding: 48 }}>
          <div className="empty-icon">🦞</div>
          <div>在上方输入关键词搜索，或点击「浏览全部」查看所有可用技能</div>
        </div>
      ) : skills.length === 0 ? (
        <div className="card empty" style={{ padding: 48 }}>
          <div className="empty-icon">🔍</div>
          <div>未找到匹配的技能</div>
          <div style={{ fontSize: 12, color: "var(--fg-dim)", marginTop: 8 }}>
            试试其他关键词，或点击「浏览全部」查看完整列表
          </div>
        </div>
      ) : (
        <>
          <div style={{ fontSize: 13, color: "var(--fg-dim)", marginBottom: 12 }}>
            找到 <strong style={{ color: "var(--fg)" }}>{skills.length}</strong> 个匹配「{keyword || "全部"}」的技能
          </div>
          <div className="clawhub-grid">
            {skills.map((s) => {
              const key = `${s.owner}/${s.slug}`;
              const st = installStates[key] || "idle";
              const err = installError[key];
              const res = installResults[key];
              return (
                <div key={key} className="clawhub-card">
                  <div className="clawhub-card-header">
                    <div className="clawhub-skill-name">
                      <span style={{ fontSize: 18 }}>📦</span>
                      <span>{s.name || s.slug}</span>
                    </div>
                    <code className="clawhub-id">{key}</code>
                  </div>
                  {s.description && (
                    <div className="clawhub-desc">{s.description}</div>
                  )}
                  <div className="clawhub-meta">
                    {s.version && <span className="meta-item">v{s.version}</span>}
                    {s.author && <span className="meta-item">👤 {s.author}</span>}
                    {s.downloads != null && <span className="meta-item">⬇ {s.downloads.toLocaleString()}</span>}
                    {s.stars != null && <span className="meta-item">★ {s.stars}</span>}
                  </div>
                  {s.tags && s.tags.length > 0 && (
                    <div className="clawhub-tags">
                      {s.tags.map((t) => (
                        <span key={t} className="tag-pill">#{t}</span>
                      ))}
                    </div>
                  )}
                  <div className="clawhub-actions">
                    <button
                      className="btn btn-primary clawhub-install-btn"
                      disabled={st === "loading" || st === "success"}
                      onClick={() => onInstallSkill(s)}
                    >
                      {st === "loading" ? "⏳ 安装中…" :
                        st === "success" ? "✅ 已安装" :
                          "⬇ 安装到 Agent"}
                    </button>
                    {res && (
                      <span className="install-result" style={{ color: "var(--success)" }}>
                        {res.skill_name} {res.version && `@ ${res.version}`}
                      </span>
                    )}
                  </div>
                  {err && (
                    <div className="install-error">❌ {err}</div>
                  )}
                </div>
              );
            })}
          </div>
        </>
      )}

      <style>{styles}</style>
    </div>
  );
}

function RuntimeBox({ label, ok, version }: { label: string; ok?: boolean; version?: string }) {
  return (
    <div className="runtime-box">
      <div className="runtime-label">{label}</div>
      <div className={`runtime-value ${ok ? "ok" : "missing"}`}>
        {ok ? (
          <>已安装{version && <span style={{ marginLeft: 4, fontSize: 11, opacity: 0.7 }}>{version}</span>}</>
        ) : (
          "未检测到"
        )}
      </div>
    </div>
  );
}

const styles = `
.runtime-box {
  padding: 10px 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
}
.runtime-label {
  font-size: 11px;
  color: var(--fg-dim);
  margin-bottom: 4px;
}
.runtime-value {
  font-size: 13px;
  font-weight: 600;
}
.runtime-value.ok { color: var(--success); }
.runtime-value.missing { color: var(--fg-dim); }

.clawhub-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 14px;
}
.clawhub-card {
  background: var(--bg-2);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  transition: border-color .15s;
}
.clawhub-card:hover {
  border-color: var(--accent);
}
.clawhub-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.clawhub-skill-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  font-size: 15px;
}
.clawhub-id {
  font-size: 11px;
  background: var(--bg);
  padding: 2px 6px;
  border-radius: 4px;
  color: var(--fg-dim);
}
.clawhub-desc {
  font-size: 13px;
  color: var(--fg);
  opacity: 0.85;
  line-height: 1.5;
}
.clawhub-meta {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  font-size: 12px;
  color: var(--fg-dim);
}
.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}
.clawhub-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.tag-pill {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--bg);
  color: var(--fg-dim);
  border: 1px solid var(--border);
}
.clawhub-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 4px;
}
.clawhub-install-btn {
  flex-shrink: 0;
}
.install-result {
  font-size: 12px;
}
.install-error {
  font-size: 12px;
  color: var(--danger);
  padding: 6px 10px;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 4px;
  word-break: break-all;
}
`;
