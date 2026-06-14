import { useState, useEffect, useCallback } from "react";
import { ChevronDown, ChevronRight, CheckSquare, Square, RotateCcw } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "./ui/card";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import { listAgents } from "../bridge";
import type { AgentInfo, AgentGroup } from "../types";

const STORAGE_KEY = "lastAgentSelection";

interface Props {
  /** Called when the set of selected agent IDs changes */
  onSelectionChange: (selected: string[]) => void;
  /** Initial selection to restore */
  initialSelection?: string[];
}

function groupAgentsByPath(agents: AgentInfo[]): AgentGroup[] {
  const groups = new Map<string, AgentGroup>();

  const detected = agents.filter((a) => a.detected);
  const undetected = agents.filter((a) => !a.detected);

  for (const a of [...detected, ...undetected]) {
    const key = a.path;
    if (!groups.has(key)) {
      groups.set(key, {
        path: key,
        agents: [{ id: a.id, name: a.name }],
        detected: a.detected,
        displayName: a.name,
        tooltipName: a.name,
      });
    } else {
      const group = groups.get(key)!;
      group.agents.push({ id: a.id, name: a.name });
      if (a.detected) group.detected = true;
    }
  }

  for (const group of groups.values()) {
    const sorted = [...group.agents].sort((a, b) => a.name.length - b.name.length);
    const shortestName = sorted[0]?.name ?? group.agents[0]?.name ?? "";
    group.displayName = group.detected ? shortestName : `${shortestName}等`;
    group.tooltipName = group.agents.map((a) => a.name).join(", ");
  }

  return Array.from(groups.values());
}

function loadSavedSelection(): Set<string> | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed)) {
        return new Set(parsed);
      }
    }
  } catch {
    // Ignore corrupt localStorage
  }
  return null;
}

function saveSelection(ids: string[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(ids));
  } catch {
    // Silently fail if localStorage is unavailable
  }
}

export default function AgentSelector({ onSelectionChange, initialSelection }: Props) {
  const [open, setOpen] = useState(true);
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [restored, setRestored] = useState(false);

  // Fetch agents on mount
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    listAgents()
      .then((data) => {
        if (!cancelled) {
          setAgents(data);
          setLoading(false);
        }
      })
      .catch(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  // Restore selection from localStorage + initialSelection
  useEffect(() => {
    if (restored || loading) return;

    let ids: string[] | null = null;
    const saved = loadSavedSelection();
    if (saved && saved.size > 0) {
      // Filter to only include IDs that exist in the current agent list
      const validIds = [...saved].filter((id) => agents.some((a) => a.id === id));
      if (validIds.length > 0) {
        ids = validIds;
      }
    }

    if (!ids && initialSelection && initialSelection.length > 0) {
      ids = initialSelection;
    }

    if (ids) {
      setSelectedIds(new Set(ids));
      onSelectionChange(ids);
    }

    setRestored(true);
  }, [agents, loading, restored, initialSelection, onSelectionChange]);

  const groups = groupAgentsByPath(agents);

  const toggleGroup = useCallback(
    (group: AgentGroup) => {
      setSelectedIds((prev) => {
        const allSelected = group.agents.every((a) => prev.has(a.id));
        const next = new Set(prev);
        for (const a of group.agents) {
          if (allSelected) {
            next.delete(a.id);
          } else {
            next.add(a.id);
          }
        }
        saveSelection([...next]);
        onSelectionChange([...next]);
        return next;
      });
    },
    [onSelectionChange],
  );

  const handleSelectAll = useCallback(() => {
    const allIds = agents.map((a) => a.id);
    const next = new Set(allIds);
    setSelectedIds(next);
    saveSelection(allIds);
    onSelectionChange(allIds);
  }, [agents, onSelectionChange]);

  const handleClear = useCallback(() => {
    setSelectedIds(new Set());
    saveSelection([]);
    onSelectionChange([]);
  }, [onSelectionChange]);

  const handleRestoreLast = useCallback(() => {
    const saved = loadSavedSelection();
    if (saved && saved.size > 0) {
      const validIds = [...saved].filter((id) => agents.some((a) => a.id === id));
      const next = new Set(validIds);
      setSelectedIds(next);
      saveSelection(validIds);
      onSelectionChange(validIds);
    }
  }, [agents, onSelectionChange]);

  const detectedCount = agents.filter((a) => a.detected).length;
  const totalCount = agents.length;

  return (
    <Card>
      <CardHeader className="pb-2 cursor-pointer" onClick={() => setOpen(!open)}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {open ? <ChevronDown className="h-4 w-4 text-muted-foreground" /> : <ChevronRight className="h-4 w-4 text-muted-foreground" />}
            <CardTitle className="text-sm">安装目标 Agent</CardTitle>
            <Badge variant="outline" className="text-[10px]">
              {loading ? "..." : `${selectedIds.size}/${totalCount}`}
            </Badge>
          </div>
          <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
            <Button variant="ghost" size="sm" className="h-6 text-xs px-2" onClick={handleSelectAll} title="全选">
              <CheckSquare className="h-3 w-3 mr-1" />
              全选
            </Button>
            <Button variant="ghost" size="sm" className="h-6 text-xs px-2" onClick={handleClear} title="清除选择">
              <Square className="h-3 w-3 mr-1" />
              清除
            </Button>
            <Button variant="ghost" size="sm" className="h-6 text-xs px-2" onClick={handleRestoreLast} title="恢复上次选择">
              <RotateCcw className="h-3 w-3 mr-1" />
              上次
            </Button>
          </div>
        </div>
      </CardHeader>

      {open && (
        <CardContent>
          {loading ? (
            <p className="text-xs text-muted-foreground">加载 Agent 列表...</p>
          ) : agents.length === 0 ? (
            <p className="text-xs text-muted-foreground">未检测到任何 Agent</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {groups.map((group) => {
                const groupSelected = group.agents.every((a) => selectedIds.has(a.id));
                const partialSelected = group.agents.some((a) => selectedIds.has(a.id)) && !groupSelected;
                return (
                  <button
                    key={group.path}
                    type="button"
                    onClick={() => toggleGroup(group)}
                    title={group.tooltipName}
                    className={`
                      inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs border transition-colors
                      ${groupSelected
                        ? "bg-primary/10 border-primary/30 text-primary"
                        : partialSelected
                          ? "bg-amber-50 border-amber-200 text-amber-700"
                          : "bg-background border-border text-muted-foreground hover:border-muted-foreground"
                      }
                      ${group.detected ? "" : "opacity-60"}
                    `}
                  >
                    <span className={`w-1.5 h-1.5 rounded-full ${group.detected ? "bg-green-500" : "bg-gray-300"}`} />
                    <span className="truncate max-w-[100px]">{group.displayName}</span>
                    {group.agents.length > 1 && (
                      <span className="text-[10px] opacity-60">x{group.agents.length}</span>
                    )}
                  </button>
                );
              })}
            </div>
          )}

          <div className="mt-2 text-[10px] text-muted-foreground">
            检测到 {detectedCount}/{totalCount} 个 Agent | 选择后将自动记住上次选择
          </div>
        </CardContent>
      )}
    </Card>
  );
}