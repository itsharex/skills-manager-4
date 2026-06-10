import { useState } from "react";
import { selectDirectory, selectFile } from "../bridge";

interface Props {
  value: string;
  onChange: (val: string) => void;
  placeholder?: string;
  /** 选择模式：directory=选择目录，file=选择文件 */
  mode?: "directory" | "file";
  /** 选择对话框标题 */
  dialogTitle?: string;
  /** 是否禁用 */
  disabled?: boolean;
  /** label 标签，可选 */
  label?: string;
  /** 提示信息 */
  hint?: string;
}

/**
 * 通用路径输入组件：支持手动输入 + 系统选择对话框
 */
export default function PathInput({
  value,
  onChange,
  placeholder = "输入路径或点击右侧选择...",
  mode = "directory",
  dialogTitle,
  disabled = false,
  label,
  hint,
}: Props) {
  const [picking, setPicking] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handlePick = async () => {
    if (disabled || picking) return;
    setPicking(true);
    setError(null);
    try {
      const title = dialogTitle || (mode === "directory" ? "选择目录" : "选择文件");
      const result = mode === "directory" ? await selectDirectory(title) : await selectFile(title);
      if (result && result.trim() !== "") {
        onChange(result);
      }
    } catch (err: any) {
      setError(String(err?.message || err));
    } finally {
      setPicking(false);
    }
  };

  return (
    <div className="path-input">
      {label && <label className="form-label">{label}</label>}
      <div className="path-input-row">
        <input
          className="form-input"
          placeholder={placeholder}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          spellCheck={false}
        />
        <button
          type="button"
          className="btn btn-secondary path-input-btn"
          onClick={handlePick}
          disabled={disabled || picking}
          title={mode === "directory" ? "选择目录" : "选择文件"}
        >
          {picking ? "⏳ 选择中..." : mode === "directory" ? "📁 选择目录" : "📄 选择文件"}
        </button>
      </div>
      {hint && <div className="form-hint">{hint}</div>}
      {error && <div className="form-hint" style={{ color: "var(--danger)" }}>{error}</div>}
    </div>
  );
}
