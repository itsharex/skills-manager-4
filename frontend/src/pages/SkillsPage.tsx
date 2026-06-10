import { useState, useMemo, useEffect } from "react";
import type { SkillWithStatus, ProjectSkill, Agent, TagUsage } from "../types";
import {
  listSkillsWithStatus,
  getAllTags,
  scanProjectSkills,
  migrateProjectSkillToLibrary,
  addSkillTag,
  removeSkillTag,
  installSkillToAgent,
  uninstallSkillFromAgent,
} from "../bridge";

// 状态过滤选项
type StatusFilter = "all" | "installed" | "partially_installed" | "not_installed";

interface Props {
  agents: Record<string, Agent>;
  onRefresh: () => void;
}

export default function SkillsPage({ agents, onRefresh }: Props) {
  // ====== 全局技能库 ======
  const [skills, setSkills] = useState<SkillWithStatus[]>([]);
  const [tags, setTags] = useState<TagUsage[]>([]);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [tagFilter, setTagFilter] = useState<string>("");
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);

  // ====== 项目技能 ======
  const [projectPath, setProjectPath] = useState<string>(
    localStorage.getItem("skills_project_path") || ""
  );
  const [projectSkills, setProjectSkills] = useState<ProjectSkill[]>([]);
  const [projectLoading, setProjectLoading] = useState(false);
  const [projectMsg, setProjectMsg] = useState<string>("");

  // ====== 通用 UI 状态 ======
  const [editingTagSkill, setEditingTagSkill] = useState<string | null>(null);
  const [tagInput, setTagInput] = useState("");
  const [busy, setBusy] = useState<Record<string, boolean>>({});

  // 加载全局技能库
  const loadGlobal = async () => {
    setLoading(true);
    try {
      const [s, t] = await Promise.all([listSkillsWithStatus(), getAllTags()]);
      // 防御性：确保返回的是数组，避免 undefined 导致白屏
      setSkills(Array.isArray(s) ? s : []);
      setTags(Array.isArray(t) ? t : []);
    } catch (err) {
      console.error("loadGlobal", err);
      // 出错时也设置为空数组，防止白屏
      setSkills([]);
      setTags([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadGlobal();
  }, []);

  // 过滤
  const filteredSkills = useMemo(() => {
    const q = search.trim().toLowerCase();
    return skills.filter((s) => {
      if (statusFilter !== "all" && s.installStatus !== statusFilter) return false;
      if (tagFilter && !(s.tags || []).includes(tagFilter)) return false;
      if (q) {
        const hay = `${s.name} ${s.description} ${(s.tags || []).join(" ")}`.toLowerCase();
        if (!hay.includes(q)) return false;
      }
      return true;
    });
  }, [skills, statusFilter, tagFilter, search]);

  // 状态统计
  const statusCounts = useMemo(() => {
    const c = { installed: 0, partially_installed: 0, not_installed: 0, all: 0 };
    skills.forEach((s) => {
      c.all++;
      if (s.installStatus === "installed") c.installed++;
      else if (s.installStatus === "partially_installed") c.partially_installed++;
      else c.not_installed++;
    });
    return c;
  }, [skills]);

  // 扫描项目技能
  const handleScanProject = async () => {
    if (!projectPath.trim()) {
      setProjectMsg("请输入项目目录路径");
      return;
    }
    setProjectLoading(true);
    setProjectMsg("");
    try {
      localStorage.setItem("skills_project_path", projectPath.trim());
      const r = await scanProjectSkills(projectPath.trim());
      setProjectSkills(r || []);
      setProjectMsg(`找到 ${r?.length || 0} 个项目技能`);
    } catch (err) {
      console.error(err);
      setProjectMsg("扫描失败");
    } finally {
      setProjectLoading(false);
    }
  };

  // 迁移项目技能到库
  const handleMigrate = async (skill: ProjectSkill) => {
    const key = `migrate:${skill.name}`;
    setBusy((b) => ({ ...b, [key]: true }));
    try {
      const r = await migrateProjectSkillToLibrary(skill.path, projectPath);
      if (r?.success) {
        setProjectMsg(`已迁移: ${skill.name}`);
        // 刷新
        await loadGlobal();
        await handleScanProject();
      } else {
        setProjectMsg(`迁移失败: ${r?.error || "unknown"}`);
      }
    } catch (err) {
      console.error(err);
      setProjectMsg("迁移失败");
    } finally {
      setBusy((b) => ({ ...b, [key]: false }));
    }
  };

  // 添加标签
  const handleAddTag = async (skillName: string) => {
    const tag = tagInput.trim();
    if (!tag) return;
    const ok = await addSkillTag(skillName, tag);
    if (ok) {
      setSkills((prev) =>
        prev.map((s) =>
          s.name === skillName
            ? { ...s, tags: Array.from(new Set([...(s.tags || []), tag])) }
            : s
        )
      );
      setTags((prev) => {
        const found = prev.find((t) => t.tag === tag);
        if (found) {
          return prev
            .map((t) => (t.tag === tag ? { ...t, count: t.count + 1 } : t))
            .sort((a, b) => b.count - a.count);
        }
        return [...prev, { tag, count: 1 }];
      });
    }
    setTagInput("");
    setEditingTagSkill(null);
  };

  // 移除标签
  const handleRemoveTag = async (skillName: string, tag: string) => {
    const ok = await removeSkillTag(skillName, tag);
    if (ok) {
      setSkills((prev) =>
        prev.map((s) =>
          s.name === skillName
            ? { ...s, tags: (s.tags || []).filter((t) => t !== tag) }
            : s
        )
      );
      setTags((prev) =>
        prev
          .map((t) => (t.tag === tag ? { ...t, count: Math.max(0, t.count - 1) } : t))
          .filter((t) => t.count > 0)
          .sort((a, b) => b.count - a.count)
      );
    }
  };

  // 安装 / 卸载（单个技能到某个 Agent）
  const handleInstall = async (skillName: string, agentID: string) => {
    const key = `install:${skillName}:${agentID}`;
    setBusy((b) => ({ ...b, [key]: true }));
    try {
      const ok = await installSkillToAgent(skillName, agentID);
      if (ok) {
        setSkills((prev) =>
          prev.map((s) =>
            s.name === skillName
              ? {
                  ...s,
                  installedAgents: Array.from(new Set([...(s.installedAgents || []), agentID])),
                  installStatus:
                    Array.from(new Set([...(s.installedAgents || []), agentID])).length >=
                    s.totalAgents
                      ? "installed"
                      : "partially_installed",
                }
              : s
          )
        );
      }
    } catch (err) {
      console.error(err);
    } finally {
      setBusy((b) => ({ ...b, [key]: false }));
    }
  };

  const handleUninstall = async (skillName: string, agentID: string) => {
    const key = `uninstall:${skillName}:${agentID}`;
    setBusy((b) => ({ ...b, [key]: true }));
    try {
      const ok = await uninstallSkillFromAgent(skillName, agentID);
      if (ok) {
        setSkills((prev) =>
          prev.map((s) =>
            s.name === skillName
              ? {
                  ...s,
                  installedAgents: (s.installedAgents || []).filter((a) => a !== agentID),
                  installStatus:
                    (s.installedAgents || []).filter((a) => a !== agentID).length === 0
                      ? "not_installed"
                      : "partially_installed",
                }
              : s
          )
        );
      }
    } catch (err) {
      console.error(err);
    } finally {
      setBusy((b) => ({ ...b, [key]: false }));
    }
  };

  return (
    <div className="skills-page">
      {/* ============ 顶部操作区 ============ */}
      <header className="page-header">
        <h1 className="page-title">技能管理</h1>
        <p className="page-desc">
          全局技能库共 {skills.length} 个，已安装到所有 Agent 的有 {statusCounts.installed} 个，
          部分安装 {statusCounts.partially_installed} 个，未安装 {statusCounts.not_installed} 个。
        </p>
        <div className="top-actions">
          <button className="btn btn-secondary" onClick={onRefresh}>
            🔄 刷新
          </button>
        </div>
      </header>

      {/* ============ 全局技能库：过滤区 ============ */}
      <section className="card filter-bar">
        <div className="filter-row">
          <div className="filter-label">状态</div>
          <div className="chip-group">
            {(
              [
                { k: "all", label: `全部 (${statusCounts.all})` },
                { k: "installed", label: `已安装 (${statusCounts.installed})` },
                { k: "partially_installed", label: `部分安装 (${statusCounts.partially_installed})` },
                { k: "not_installed", label: `未安装 (${statusCounts.not_installed})` },
              ] as { k: StatusFilter; label: string }[]
            ).map((opt) => (
              <button
                key={opt.k}
                className={`chip ${statusFilter === opt.k ? "active" : ""}`}
                onClick={() => setStatusFilter(opt.k)}
                type="button"
              >
                {opt.label}
              </button>
            ))}
            {(statusFilter !== "all" || tagFilter !== "" || search.trim() !== "") && (
              <button
                className="chip chip-clear"
                onClick={() => { setStatusFilter("all"); setTagFilter(""); setSearch(""); }}
                type="button"
                title="清除所有过滤条件"
              >
                ✕ 清除过滤
              </button>
            )}
          </div>
        </div>
        <div className="filter-row">
          <div className="filter-label">标签</div>
          <div className="chip-group">
            <button
              className={`chip ${tagFilter === "" ? "active" : ""}`}
              onClick={() => setTagFilter("")}
              type="button"
            >
              全部
            </button>
            {tags.map((t) => (
              <button
                key={t.tag}
                className={`chip ${tagFilter === t.tag ? "active" : ""}`}
                onClick={() => setTagFilter(tagFilter === t.tag ? "" : t.tag)}
                type="button"
              >
                #{t.tag} ({t.count})
              </button>
            ))}
          </div>
        </div>
        <div className="filter-row">
          <div className="filter-label">搜索</div>
          <input
            className="search-input"
            placeholder="按名称 / 描述 / 标签搜索..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            style={{ flex: 1, maxWidth: 420 }}
          />
        </div>
        {(statusFilter !== "all" || tagFilter !== "" || search.trim() !== "") && (
          <div style={{ fontSize: 12, color: "var(--fg-dim)" }}>
            当前过滤：{filteredSkills.length} / {skills.length} 个技能
          </div>
        )}
      </section>

      {/* ============ 全局技能库：列表 ============ */}
      {loading ? (
        <div className="card empty">
          <div className="empty-icon">⏳</div>
          <div>加载技能中...</div>
        </div>
      ) : filteredSkills.length === 0 ? (
        <div className="card empty">
          <div className="empty-icon">📦</div>
          <div>暂无匹配的技能，前往「安装技能」添加或调整过滤条件。</div>
        </div>
      ) : (
        <div className="skill-grid">
          {filteredSkills.map((s) => (
            <div key={s.name} className="skill-card">
              <div className="skill-card-head">
                <div className="skill-name">{s.name}</div>
                <span
                  className={`status-badge status-${s.installStatus}`}
                  title={s.installStatus}
                >
                  {statusLabel(s.installStatus)}
                </span>
              </div>
              <div className="skill-desc">{s.description || "—"}</div>

              {/* 标签展示 + 编辑 */}
              <div className="skill-tags">
                {(s.tags || []).map((t) => (
                  <span key={t} className="tag-chip" title={`标签：${t}`}>
                    #{t}
                    <button
                      className="tag-remove"
                      onClick={() => handleRemoveTag(s.name, t)}
                      title="移除标签"
                    >
                      ×
                    </button>
                  </span>
                ))}
                {editingTagSkill === s.name ? (
                  <span className="tag-edit">
                    <input
                      autoFocus
                      value={tagInput}
                      onChange={(e) => setTagInput(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") handleAddTag(s.name);
                        if (e.key === "Escape") {
                          setEditingTagSkill(null);
                          setTagInput("");
                        }
                      }}
                      placeholder="新标签..."
                      className="tag-input"
                    />
                    <button
                      className="btn btn-xs"
                      onClick={() => handleAddTag(s.name)}
                    >
                      添加
                    </button>
                  </span>
                ) : (
                  <button
                    className="tag-add"
                    onClick={() => {
                      setEditingTagSkill(s.name);
                      setTagInput("");
                    }}
                  >
                    + 标签
                  </button>
                )}
              </div>

              {/* 元信息 */}
              <div className="skill-meta">
                <span className="badge">v {s.latestVersion || "—"}</span>
                <span className="badge">
                  {(s.installedAgents || []).length}/{s.totalAgents} 个 Agent
                </span>
              </div>

              {/* 已安装的 Agent 列表 */}
              <div className="skill-agents">
                {Object.keys(agents).length === 0 ? (
                  <span style={{ color: "var(--fg-dim)" }}>未配置 Agent</span>
                ) : (
                  Object.entries(agents).map(([agentID, ag]) => {
                    const installed = (s.installedAgents || []).includes(agentID);
                    const key = `${installed ? "uninstall" : "install"}:${s.name}:${agentID}`;
                    const isBusy = !!busy[key];
                    return (
                      <button
                        key={agentID}
                        disabled={isBusy || !ag.detected}
                        className={`agent-btn ${installed ? "installed" : ""} ${
                          !ag.detected ? "disabled" : ""
                        }`}
                        onClick={() =>
                          installed ? handleUninstall(s.name, agentID) : handleInstall(s.name, agentID)
                        }
                        title={
                          ag.detected
                            ? installed
                              ? `从 ${ag.name} 卸载`
                              : `安装到 ${ag.name}`
                            : `${ag.name} 未检测到`
                        }
                      >
                        {isBusy ? "⏳" : installed ? "✓" : "+"} {ag.name}
                      </button>
                    );
                  })
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ============ 项目技能 ============ */}
      <section className="card project-skills-section" style={{ marginTop: 28 }}>
        <div className="section-head">
          <h2 className="page-subtitle">📂 项目技能（可选）</h2>
          <div className="section-sub">
            扫描特定项目目录下的 <code>.*/skills</code> 目录，发现本地的项目专属技能，
            可一键迁移到全局技能库（复制，不使用软链接，保证独立）。
          </div>
        </div>

        <div className="project-path-row">
          <input
            className="search-input"
            placeholder="项目目录绝对路径（例如 /Users/me/proj 或 C:\code\proj）"
            value={projectPath}
            onChange={(e) => setProjectPath(e.target.value)}
            style={{ flex: 1 }}
          />
          <button
            className="btn btn-primary"
            onClick={handleScanProject}
            disabled={projectLoading}
          >
            {projectLoading ? "扫描中..." : "扫描项目技能"}
          </button>
        </div>
        {projectMsg && <div className="project-msg">{projectMsg}</div>}

        {projectSkills.length > 0 && (
          <div className="skill-grid">
            {projectSkills.map((ps) => {
              const key = `migrate:${ps.name}`;
              const isBusy = !!busy[key];
              return (
                <div key={`${ps.path}-${ps.name}`} className="skill-card">
                  <div className="skill-card-head">
                    <div className="skill-name">{ps.name}</div>
                    <span
                      className={`status-badge ${ps.inLibrary ? "status-installed" : "status-project"}`}
                    >
                      {ps.inLibrary ? "已在库中" : "仅项目"}
                    </span>
                  </div>
                  <div className="skill-desc">{ps.description || "—"}</div>
                  {(ps.tags && ps.tags.length > 0) && (
                    <div className="skill-tags">
                      {ps.tags.map((t) => (
                        <span key={t} className="tag-chip">
                          #{t}
                        </span>
                      ))}
                    </div>
                  )}
                  <div className="skill-meta">
                    <span className="badge">v {ps.version || "—"}</span>
                    <span className="badge">
                      {ps.isSymlink ? "软链接" : "独立目录"}
                    </span>
                  </div>
                  <div className="skill-path" title={ps.path}>
                    {ps.path}
                  </div>
                  <div className="skill-actions">
                    {!ps.inLibrary ? (
                      <button
                        className="btn btn-primary"
                        onClick={() => handleMigrate(ps)}
                        disabled={isBusy}
                      >
                        {isBusy ? "迁移中..." : "迁移到全局技能库"}
                      </button>
                    ) : (
                      <span className="badge" style={{ background: "var(--success)" }}>
                        已存在于全局技能库
                      </span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>

      <style>{styles}</style>
    </div>
  );
}

function statusLabel(s: string): string {
  switch (s) {
    case "installed":
      return "已安装";
    case "partially_installed":
      return "部分安装";
    case "not_installed":
      return "未安装";
    case "project_only":
      return "仅项目";
    default:
      return s;
  }
}

// 样式
const styles = `
.skills-page { width: 100%; }
.top-actions { display: flex; gap: 8px; margin-top: 10px; }

.filter-bar { display: flex; flex-direction: column; gap: 12px; }
.filter-row { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.filter-label { width: 56px; color: var(--fg-dim); font-size: 13px; }
.chip-group { display: flex; gap: 6px; flex-wrap: wrap; }
.chip {
  padding: 6px 14px;
  border-radius: 999px;
  border: 1.5px solid var(--border);
  background: var(--bg);
  color: var(--fg);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all .15s;
}
.chip:hover {
  background: var(--bg-3);
  border-color: var(--accent);
}
.chip.active {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
  font-weight: 600;
}
.chip-clear {
  border-color: var(--danger);
  color: var(--danger);
  background: rgba(239, 68, 68, 0.08);
}
.chip-clear:hover {
  background: var(--danger);
  color: white;
}

.skill-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 14px;
  margin-top: 14px;
}
.skill-card {
  background: var(--bg-2);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.skill-card-head { display: flex; justify-content: space-between; align-items: center; }
.skill-name { font-weight: 600; font-size: 15px; }
.skill-desc { color: var(--fg-dim); font-size: 13px; min-height: 36px; }

.skill-meta { display: flex; gap: 6px; flex-wrap: wrap; }
.badge {
  display: inline-block;
  padding: 3px 8px;
  border-radius: 4px;
  background: var(--bg-3);
  color: var(--fg);
  font-size: 12px;
}
.badge.latest { background: var(--primary); color: white; }

.status-badge {
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
}
.status-installed { background: var(--success); color: white; }
.status-partially_installed { background: #e9a23b; color: white; }
.status-not_installed { background: var(--bg-3); color: var(--fg); }
.status-project { background: #6a5acd; color: white; }

.skill-tags { display: flex; gap: 6px; flex-wrap: wrap; align-items: center; }
.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  background: var(--bg-3);
  border-radius: 999px;
  font-size: 12px;
}
.tag-remove {
  background: transparent;
  border: none;
  color: var(--fg-dim);
  cursor: pointer;
  padding: 0 2px;
  font-size: 14px;
  line-height: 1;
}
.tag-remove:hover { color: #ff6b6b; }
.tag-add {
  background: transparent;
  border: 1px dashed var(--border);
  color: var(--fg-dim);
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  cursor: pointer;
}
.tag-add:hover { color: var(--fg); border-color: var(--fg-dim); }
.tag-edit { display: inline-flex; gap: 4px; align-items: center; }
.tag-input {
  width: 110px;
  padding: 2px 8px;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--bg);
  color: var(--fg);
  font-size: 12px;
}

.skill-agents { display: flex; gap: 6px; flex-wrap: wrap; }
.agent-btn {
  padding: 4px 10px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--fg);
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  transition: all .15s;
}
.agent-btn:hover:not(.disabled) { background: var(--bg-3); }
.agent-btn.installed {
  background: var(--success);
  color: white;
  border-color: var(--success);
}
.agent-btn.disabled { opacity: 0.4; cursor: not-allowed; }

.project-skills-section .section-head { margin-bottom: 14px; }
.project-skills-section .section-sub { color: var(--fg-dim); font-size: 13px; margin-top: 6px; }
.project-skills-section .project-path-row {
  display: flex; gap: 8px; margin-bottom: 10px;
}
.project-msg { color: var(--fg-dim); font-size: 13px; margin-bottom: 8px; }
.skill-path {
  font-family: ui-monospace, monospace;
  font-size: 11px;
  color: var(--fg-dim);
  word-break: break-all;
  background: var(--bg);
  padding: 4px 6px;
  border-radius: 4px;
}
.skill-actions { display: flex; gap: 6px; }
.btn-xs { padding: 2px 8px; font-size: 12px; }
`;
