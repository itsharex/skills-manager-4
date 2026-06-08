import { useState } from "react";
import type { Agent, InstallRequest, InstallResult } from "../types";
import { installSkill } from "../bridge";

interface Props {
  agents: Record<string, Agent>;
  onInstalled: () => void;
}

export default function InstallPage({ agents, onInstalled }: Props) {
  const [source, setSource] = useState("");
  const [subPath, setSubPath] = useState("");
  const [version, setVersion] = useState("");
  const [ref, setRef] = useState("");
  const [selectedAgents, setSelectedAgents] = useState<string[]>([]);
  const [scope, setScope] = useState<"global" | "project">("global");
  const [projectDir, setProjectDir] = useState("");
  const [loading, setLoading] = useState(false);
  const [output, setOutput] = useState<string>("");
  const [error, setError] = useState<string>("");

  // Default tick detected agents the first time they appear
  const agentIds = Object.keys(agents);
  const detectedAgents = agentIds.filter((id) => agents[id].detected);

  // Auto-select detected agents on first load
  const ensureDefaults = () => {
    if (selectedAgents.length === 0 && detectedAgents.length > 0) {
      setSelectedAgents(detectedAgents);
    }
  };
  if (agentIds.length > 0 && selectedAgents.length === 0) {
    ensureDefaults();
  }

  const toggleAgent = (id: string) => {
    setSelectedAgents((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    );
  };

  const handleInstall = async () => {
    setError("");
    if (!source.trim()) {
      setError("请填写来源（GitHub URL 或本地路径）");
      return;
    }
    if (selectedAgents.length === 0) {
      setError("请至少选择一个目标 Agent");
      return;
    }
    setLoading(true);
    setOutput("⏳ 正在解析 SKILL.md 并安装...\n");
    try {
      const req: InstallRequest = {
        source: source.trim(),
        sub_path: subPath.trim() || undefined,
        version: version.trim() || undefined,
        ref: ref.trim() || undefined,
        agents: selectedAgents,
        scope,
        project_dir: scope === "project" ? projectDir.trim() || undefined : undefined,
      };
      const results: InstallResult[] = await installSkill(req);
      let out = "";
      for (const r of results) {
        out += `✅ 已安装: ${r.skill_name} @ ${r.version}\n`;
        for (const agentId of Object.keys(r.agents)) {
          const link = r.agents[agentId];
          if (link.success) {
            out += `   ✅ ${agentId} → ${link.path}\n`;
          } else {
            out += `   ❌ ${agentId}: ${link.error}\n`;
          }
        }
        out += "\n";
      }
      setOutput(out);
      onInstalled();
    } catch (e: any) {
      setError(String(e?.message || e));
      setOutput("");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <header className="page-header">
        <h1 className="page-title">安装技能</h1>
        <p className="page-desc">
          从 GitHub 仓库或本地目录安装 SKILL.md 技能，分发到多个 Agent。
        </p>
      </header>

      <div className="card">
        <div className="form-group">
          <label className="form-label">来源（GitHub URL 或本地路径）</label>
          <input
            className="form-input"
            placeholder="https://github.com/user/repo 或 /path/to/local/skill"
            value={source}
            onChange={(e) => setSource(e.target.value)}
          />
        </div>

        <div className="form-row">
          <div className="form-group">
            <label className="form-label">子目录 (可选)</label>
            <input
              className="form-input"
              placeholder="skills/my-skill"
              value={subPath}
              onChange={(e) => setSubPath(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label className="form-label">分支 / 标签 (可选)</label>
            <input
              className="form-input"
              placeholder="main / v1.0.0"
              value={ref}
              onChange={(e) => setRef(e.target.value)}
            />
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

        <div className="form-group">
          <label className="form-label">安装范围</label>
          <div className="checkbox-group">
            <label
              className={"checkbox-item" + (scope === "global" ? " checked" : "")}
            >
              <input
                type="radio"
                name="scope"
                checked={scope === "global"}
                onChange={() => setScope("global")}
              />
              全局（所有项目可用）
            </label>
            <label
              className={"checkbox-item" + (scope === "project" ? " checked" : "")}
            >
              <input
                type="radio"
                name="scope"
                checked={scope === "project"}
                onChange={() => setScope("project")}
              />
              项目内（仅某项目）
            </label>
          </div>
        </div>

        {scope === "project" && (
          <div className="form-group">
            <label className="form-label">项目目录</label>
            <input
              className="form-input"
              placeholder="/path/to/project"
              value={projectDir}
              onChange={(e) => setProjectDir(e.target.value)}
            />
          </div>
        )}

        <div className="form-group">
          <label className="form-label">
            目标 Agent（已检测到 {detectedAgents.length} 个，共 {agentIds.length} 个）
          </label>
          <div className="checkbox-group">
            {agentIds.map((id) => {
              const a = agents[id];
              const checked = selectedAgents.includes(id);
              return (
                <label
                  key={id}
                  className={"checkbox-item" + (checked ? " checked" : "")}
                >
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={() => toggleAgent(id)}
                  />
                  <span>
                    {a.name}
                    <span
                      style={{
                        marginLeft: 8,
                        color: a.detected ? "var(--success)" : "var(--fg-dim)",
                        fontSize: 11,
                      }}
                    >
                      {a.detected ? "● 已检测" : "○ 未检测"}
                    </span>
                  </span>
                </label>
              );
            })}
          </div>
        </div>

        <button className="btn" onClick={handleInstall} disabled={loading}>
          {loading ? "安装中..." : "安装"}
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
