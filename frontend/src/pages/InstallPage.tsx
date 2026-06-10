import { useState, useEffect } from "react";
import type { AgentGroup, InstallRequest, InstallResult } from "../types";
import { installSkill, listAgentGroups } from "../bridge";
import PathInput from "../components/PathInput";

interface Props {
  onInstalled: () => void;
}

// 安装方式：必须在全局和项目之间 2 选 1
type Scope = "global" | "project";

export default function InstallPage({ onInstalled }: Props) {
  const [source, setSource] = useState("");
  const [subPath, setSubPath] = useState("");
  const [version, setVersion] = useState("");
  const [ref, setRef] = useState("");
  const [scope, setScope] = useState<Scope | "">(""); // 初始为空，强制用户选择
  const [projectDir, setProjectDir] = useState("");
  const [groups, setGroups] = useState<AgentGroup[]>([]);
  const [loading, setLoading] = useState(false);
  const [output, setOutput] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [selectedAgentIds, setSelectedAgentIds] = useState<Set<string>>(new Set());

  // 根据 scope 加载对应的目录组
  useEffect(() => {
    if (scope === "global" || scope === "project") {
      loadGroups(scope);
    } else {
      setGroups([]);
      setSelectedAgentIds(new Set());
    }
  }, [scope]);

  const loadGroups = async (s: Scope) => {
    try {
      const gs = await listAgentGroups(s);
      setGroups(gs || []);
      // 默认全选已检测的
      const ids = new Set<string>();
      for (const g of gs || []) {
        for (const id of g.detectedIds) {
          ids.add(id);
        }
      }
      setSelectedAgentIds(ids);
    } catch (e) {
      console.error("load groups error", e);
    }
  };

  const toggleAgent = (id: string) => {
    setSelectedAgentIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  // 按组全选 / 取消全选
  const toggleGroup = (g: AgentGroup) => {
    const groupAgentIds = g.agentIds;
    const allSelected = groupAgentIds.every((id) => selectedAgentIds.has(id));
    setSelectedAgentIds((prev) => {
      const next = new Set(prev);
      for (const id of groupAgentIds) {
        if (allSelected) {
          next.delete(id);
        } else {
          next.add(id);
        }
      }
      return next;
    });
  };

  const handleInstall = async () => {
    setError("");
    setOutput("");

    if (!source.trim()) {
      setError("请填写来源（GitHub URL 或本地路径）");
      return;
    }
    if (scope !== "global" && scope !== "project") {
      setError("请选择安装方式（全局或项目级）");
      return;
    }
    if (selectedAgentIds.size === 0) {
      setError("请至少选择一个目标 Agent");
      return;
    }
    if (scope === "project" && !projectDir.trim()) {
      setError("项目级安装请填写项目目录");
      return;
    }

    setLoading(true);
    try {
      const req: InstallRequest = {
        source: source.trim(),
        sub_path: subPath.trim() || undefined,
        version: version.trim() || undefined,
        ref: ref.trim() || undefined,
        agents: Array.from(selectedAgentIds),
        scope,
        project_dir: scope === "project" ? projectDir.trim() || undefined : undefined,
      };
      const results: InstallResult[] = await installSkill(req);
      let out = "";
      for (const r of results) {
        out += `✅ 已安装: ${r.skill_name} @ ${r.version}\n`;
        for (const agentId of Object.keys(r.agents || {})) {
          const link = (r.agents || {})[agentId];
          if (link?.success) {
            out += `   ✅ ${agentId} → ${link.path}\n`;
          } else {
            out += `   ❌ ${agentId}: ${link?.error || "未知错误"}\n`;
          }
        }
        out += "\n";
      }
      setOutput(out);
      onInstalled();
    } catch (e: any) {
      setError(String(e?.message || e));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <header className="page-header">
        <h1 className="page-title">安装技能</h1>
        <p className="page-desc">
          从 GitHub 仓库或本地目录安装 SKILL.md 技能，分发到多个 Agent 目录。
        </p>
      </header>

      <div className="card">
        <div className="form-group">
          <label className="form-label">来源（GitHub URL 或本地路径）</label>
          <PathInput
            value={source}
            onChange={setSource}
            mode="directory"
            dialogTitle="选择本地技能目录"
            placeholder="https://github.com/user/repo 或 /path/to/local/skill"
            hint="支持 GitHub 仓库 URL、owner/repo 简写、本地目录路径（含 SKILL.md 的目录）"
          />
        </div>

        <div className="form-row">
          <div className="form-group">
            <label className="form-label">仓库子目录 (可选)</label>
            <input
              className="form-input"
              placeholder="例如: skills/create-plan（用于仓库含多个技能目录）"
              value={subPath}
              onChange={(e) => setSubPath(e.target.value)}
            />
            <div className="form-hint">该仓库可能包含多个技能目录，通过子目录指定要安装的具体技能</div>
          </div>
          <div className="form-group">
            <label className="form-label">Git 分支或版本标签 (可选)</label>
            <input
              className="form-input"
              placeholder="例如: main、dev、v1.2.0（留空默认最新）"
              value={ref}
              onChange={(e) => setRef(e.target.value)}
            />
            <div className="form-hint">指定 Git 分支（如 main）或版本标签（如 v1.0.0）来拉取特定版本</div>
          </div>
          <div className="form-group">
            <label className="form-label">强制版本号 (可选)</label>
            <input
              className="form-input"
              placeholder="1.0.0（留空则从 SKILL.md 读取）"
              value={version}
              onChange={(e) => setVersion(e.target.value)}
            />
          </div>
        </div>

        {/* 安装方式：强制用户选择 */}
        <div className="form-group" style={{ marginTop: 16 }}>
          <label className="form-label" style={{ fontWeight: 700 }}>
            安装方式（必须选择一项）
          </label>
          <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
            <label
              className="checkbox-item"
              style={{
                padding: "12px 16px",
                border: "1px solid " + (scope === "global" ? "var(--accent)" : "var(--border)"),
                borderRadius: 8,
                cursor: "pointer",
                background: scope === "global" ? "var(--bg-alt)" : "transparent",
                display: "flex",
                alignItems: "center",
                gap: 8,
                minWidth: 200,
              }}
            >
              <input
                type="radio"
                name="scope-global"
                checked={scope === "global"}
                onChange={() => setScope("global")}
              />
              <div>
                <strong>全局安装</strong>
                <div style={{ fontSize: 12, color: "var(--fg-dim)", marginTop: 2 }}>
                  所有项目都可使用，但技能会写入用户目录
                </div>
              </div>
            </label>
            <label
              className="checkbox-item"
              style={{
                padding: "12px 16px",
                border: "1px solid " + (scope === "project" ? "var(--accent)" : "var(--border)"),
                borderRadius: 8,
                cursor: "pointer",
                background: scope === "project" ? "var(--bg-alt)" : "transparent",
                display: "flex",
                alignItems: "center",
                gap: 8,
                minWidth: 200,
              }}
            >
              <input
                type="radio"
                name="scope-project"
                checked={scope === "project"}
                onChange={() => setScope("project")}
              />
              <div>
                <strong>项目级安装</strong>
                <div style={{ fontSize: 12, color: "var(--fg-dim)", marginTop: 2 }}>
                  仅当前项目可用，不污染全局目录
                </div>
              </div>
            </label>
          </div>
        </div>

        {/* 全局安装风险提示 */}
        {scope === "global" && (
          <div
            style={{
              margin: "16px 0",
              padding: "12px 14px",
              borderRadius: 8,
              background: "rgba(255, 180, 0, 0.12)",
              border: "1px solid rgba(255, 180, 0, 0.4)",
              color: "var(--fg)",
              fontSize: 13,
              lineHeight: 1.6,
            }}
          >
            <strong>⚠️ 全局安装风险提示</strong>
            <ul style={{ margin: "8px 0 0 20px", padding: 0 }}>
              <li>技能目录会写入用户 home 目录，其他项目可能误读取</li>
              <li>
                多 Agent 共用同一目录时，可能出现<span style={{ color: "var(--accent)" }}>数据混乱</span>
                ，请确认各 Agent 目录互不冲突
              </li>
              <li>卸载时将清除全局目录，谨慎操作</li>
            </ul>
          </div>
        )}

        {/* 项目目录输入 */}
        {scope === "project" && (
          <div className="form-group">
            <label className="form-label">项目目录</label>
            <input
              className="form-input"
              placeholder="/path/to/project"
              value={projectDir}
              onChange={(e) => setProjectDir(e.target.value)}
            />
            <div style={{ fontSize: 12, color: "var(--fg-dim)", marginTop: 6 }}>
              项目级安装会把技能复制到项目根目录下的 Agent 特定目录（例如 .agents/skills 或 .claude/skills）
            </div>
          </div>
        )}

        {/* Agent 按目录组选择 */}
        {scope !== "" && (
          <div className="form-group" style={{ marginTop: 12 }}>
            <label className="form-label" style={{ fontWeight: 700 }}>
              目标 Agent（按目录分组 · 当前共 {groups.length} 组，
              已选 {selectedAgentIds.size} 个 Agent）
            </label>

            {groups.length === 0 && (
              <div
                style={{
                  padding: "12px 14px",
                  background: "var(--bg-alt)",
                  borderRadius: 8,
                  color: "var(--fg-dim)",
                  fontSize: 13,
                }}
              >
                当前没有配置的 Agent。请先在配置文件中添加 Agent。
              </div>
            )}

            <div style={{ display: "flex", flexDirection: "column", gap: 10, marginTop: 8 }}>
              {groups.map((g) => {
                const groupAllSelected = g.agentIds.every((id) => selectedAgentIds.has(id));
                const groupSomeSelected = g.agentIds.some((id) => selectedAgentIds.has(id));
                return (
                  <div
                    key={g.id}
                    style={{
                      padding: "12px 14px",
                      border:
                        "1px solid " +
                        (groupSomeSelected ? "var(--accent)" : "var(--border)"),
                      borderRadius: 8,
                      background: groupSomeSelected ? "rgba(59, 130, 246, 0.06)" : "transparent",
                    }}
                  >
                    {/* 目录组标题 */}
                    <div
                      style={{
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "space-between",
                        gap: 12,
                        marginBottom: 10,
                      }}
                    >
                      <div
                        style={{
                          display: "flex",
                          alignItems: "center",
                          gap: 8,
                          cursor: "pointer",
                        }}
                        onClick={() => toggleGroup(g)}
                      >
                        <input
                          type="checkbox"
                          checked={groupAllSelected}
                          onChange={() => toggleGroup(g)}
                          style={{ transform: "scale(1.1)", cursor: "pointer" }}
                        />
                        <strong style={{ fontSize: 13 }}>{g.directory}</strong>
                        <span style={{ fontSize: 11, color: "var(--fg-dim)" }}>
                          （{g.agentIds.length} 个 Agent，已检测 {g.detectedIds.length} 个）
                        </span>
                      </div>

                      {g.sharedRisk && (
                        <span
                          style={{
                            fontSize: 11,
                            color: "#f59e0b",
                            fontWeight: 600,
                            padding: "2px 8px",
                            background: "rgba(245, 158, 11, 0.15)",
                            borderRadius: 999,
                          }}
                        >
                          ⚠️ 目录共享风险
                        </span>
                      )}
                    </div>

                    {/* 共享目录风险提示 */}
                    {g.sharedRisk && (
                      <div
                        style={{
                          fontSize: 12,
                          color: "#f59e0b",
                          padding: "6px 10px",
                          background: "rgba(245, 158, 11, 0.08)",
                          borderRadius: 6,
                          marginBottom: 10,
                        }}
                      >
                        以下 {g.agentIds.length} 个 Agent 共用此目录，安装的技能可能被多个 Agent 同时访问：
                        <br />
                        <span style={{ color: "var(--fg-dim)" }}>
                          {g.agentNames.join("、")}
                        </span>
                      </div>
                    )}

                    {/* 具体 Agent 复选 */}
                    <div style={{ display: "flex", flexWrap: "wrap", gap: 10 }}>
                      {g.agentIds.map((id, idx) => {
                        const name = g.agentNames[idx] || id;
                        const detected = g.detectedIds.includes(id);
                        const checked = selectedAgentIds.has(id);
                        return (
                          <label
                            key={id}
                            style={{
                              display: "flex",
                              alignItems: "center",
                              gap: 6,
                              padding: "6px 10px",
                              border:
                                "1px solid " + (checked ? "var(--accent)" : "var(--border)"),
                              borderRadius: 6,
                              cursor: "pointer",
                              fontSize: 13,
                              background: checked ? "var(--bg-alt)" : "transparent",
                            }}
                          >
                            <input
                              type="checkbox"
                              checked={checked}
                              onChange={() => toggleAgent(id)}
                              style={{ cursor: "pointer" }}
                            />
                            <span>{name}</span>
                            <span
                              style={{
                                fontSize: 10,
                                color: detected ? "var(--success)" : "var(--fg-dim)",
                              }}
                            >
                              {detected ? "●已检测" : "○未检测"}
                            </span>
                          </label>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        <button
          className="btn"
          onClick={handleInstall}
          disabled={loading}
          style={{ marginTop: 20 }}
        >
          {loading ? "安装中..." : `安装到 ${selectedAgentIds.size} 个 Agent`}
        </button>
      </div>

      {error && (
        <div className="card">
          <div className="error-msg">❌ {error}</div>
        </div>
      )}

      {output && (
        <div className="card">
          <h3 className="card-title">安装结果</h3>
          <pre className="status-box success-msg">{output}</pre>
        </div>
      )}
    </div>
  );
}
