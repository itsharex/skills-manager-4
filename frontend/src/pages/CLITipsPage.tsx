import { Terminal, Command, FolderTree, Info } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import { BookOpen } from "lucide-react";

interface CLITip {
  name: string;
  command: string;
  install: string;
  skillsDir: string;
  type: "cli" | "desktop";
  description: string;
  docs?: string;
}

const CLI_TIPS: CLITip[] = [
  {
    name: "Claude Code",
    command: "claude",
    install: "npm install -g @anthropic-ai/claude-code",
    skillsDir: "~/.claude/skills/",
    type: "cli",
    description: "Anthropic 官方的 CLI Agent，支持通过 SKILL.md 加载技能。",
    docs: "https://docs.anthropic.com/en/docs/claude-code/overview",
  },
  {
    name: "OpenAI Codex CLI",
    command: "codex",
    install: "npm install -g @openai/codex-cli",
    skillsDir: "~/.codex/skills/",
    type: "cli",
    description: "OpenAI 的 CLI Agent，支持技能扩展。",
    docs: "https://github.com/openai/codex",
  },
  {
    name: "Gemini CLI",
    command: "gemini",
    install: "npm install -g @anthropic-ai/gemini-cli",
    skillsDir: "~/.gemini/skills/",
    type: "cli",
    description: "Google Gemini 的 CLI Agent。",
  },
  {
    name: "Antigravity CLI",
    command: "agy",
    install: "go install github.com/anthropics/antigravity@latest",
    skillsDir: "~/.gemini/antigravity/skills/",
    type: "cli",
    description: "Antigravity 的 CLI 版本，使用 agy 命令。",
  },
  {
    name: "Aider",
    command: "aider",
    install: "pip install aider-chat",
    skillsDir: "~/.aider/skills/",
    type: "cli",
    description: "Aider 是 AI 驱动的代码编辑工具。",
    docs: "https://aider.chat/",
  },
  {
    name: "OpenCode",
    command: "opencode",
    install: "go install github.com/opencode-ai/opencode@latest",
    skillsDir: "~/.config/opencode/skills/",
    type: "cli",
    description: "OpenCode 是开源的 CLI Agent。",
    docs: "https://github.com/opencode-ai/opencode",
  },
  {
    name: "Ollama",
    command: "ollama",
    install: "brew install ollama",
    skillsDir: "~/.ollama/skills/",
    type: "cli",
    description: "本地运行大语言模型的 CLI 工具。",
    docs: "https://ollama.com/",
  },
  {
    name: "Goose",
    command: "goose",
    install: "brew install tunnelbear/goose/goose",
    skillsDir: "~/.goose/skills/",
    type: "cli",
    description: "Anthropic 的开源 Agent 工具。",
    docs: "https://github.com/block/goose",
  },
];

export default function CLITipsPage() {
  const cliTips = CLI_TIPS.filter(t => t.type === "cli");
  const desktopTips = CLI_TIPS.filter(t => t.type === "desktop");

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight flex items-center gap-3">
          <Terminal className="h-6 w-6 text-primary" />
          CLI 命令教程
        </h2>
        <p className="text-muted-foreground mt-1">
          常见 CLI Agent 的安装命令、使用方式和技能目录路径
        </p>
      </div>

      {/* CLI Agents */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-sm">
            <Command className="h-4 w-4" />
            CLI Agent ({cliTips.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-muted/50 border-b">
                  <th className="text-left font-medium px-4 py-2">智能体工具</th>
                  <th className="text-left font-medium px-4 py-2">检测命令</th>
                  <th className="text-left font-medium px-4 py-2">安装方式</th>
                  <th className="text-left font-medium px-4 py-2">技能目录</th>
                  <th className="text-left font-medium px-4 py-2">说明</th>
                </tr>
              </thead>
              <tbody>
                {cliTips.map((tip) => (
                  <tr key={tip.name} className={`border-b last:border-0`}>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <BookOpen className="h-4 w-4 text-muted-foreground" />
                        <span className="font-medium">{tip.name}</span>
                        <Badge variant="outline" className="text-[10px]">CLI</Badge>
                      </div>
                    </td>
                    <td className="px-4 py-3 font-mono text-xs">{tip.command}</td>
                    <td className="px-4 py-3">
                      <code className="text-xs bg-muted px-2 py-1 rounded">{tip.install}</code>
                    </td>
                    <td className="px-4 py-3">
                      <code className="text-xs bg-muted px-2 py-1 rounded flex items-center gap-1">
                        <FolderTree className="h-3 w-3" />
                        {tip.skillsDir}
                      </code>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground max-w-xs truncate">
                      {tip.description}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Desktop Agents */}
      {desktopTips.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-sm">
              <FolderTree className="h-4 w-4" />
              桌面 Agent ({desktopTips.length})
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="border rounded-lg overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-muted/50 border-b">
                    <th className="text-left font-medium px-4 py-2">智能体工具</th>
                    <th className="text-left font-medium px-4 py-2">安装方式</th>
                    <th className="text-left font-medium px-4 py-2">技能目录</th>
                    <th className="text-left font-medium px-4 py-2">说明</th>
                  </tr>
                </thead>
                <tbody>
                  {desktopTips.map((tip) => (
                    <tr key={tip.name} className={`border-b last:border-0`}>
                      <td className="px-4 py-3 font-medium flex items-center gap-2">
                        {tip.name}
                        <Badge variant="outline" className="text-[10px]">Desktop</Badge>
                      </td>
                      <td className="px-4 py-3">
                        <code className="text-xs bg-muted px-2 py-1 rounded">{tip.install}</code>
                      </td>
                      <td className="px-4 py-3">
                        <code className="text-xs bg-muted px-2 py-1 rounded">{tip.skillsDir}</code>
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">{tip.description}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Quick Start */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-sm">
            <Info className="h-4 w-4 text-blue-500" />
            快速入门
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm">
          <div>
            <p className="font-medium mb-1">1. 安装 CLI Agent</p>
            <p className="text-muted-foreground">选择上方表格中的命令，在终端中执行安装。</p>
          </div>
          <div>
            <p className="font-medium mb-1">2. 创建技能目录</p>
            <p className="text-muted-foreground">
              根据技能目录路径创建目录，例如：<code className="text-xs bg-muted px-1 rounded">mkdir -p ~/.claude/skills/</code>
            </p>
          </div>
          <div>
            <p className="font-medium mb-1">3. 添加 SKILL.md</p>
            <p className="text-muted-foreground">
              在技能目录中添加 <code className="text-xs bg-muted px-1 rounded">SKILL.md</code> 文件定义你的技能。
            </p>
          </div>
          <div>
            <p className="font-medium mb-1">4. 使用 Skills Manager</p>
            <p className="text-muted-foreground">
              打开 Skills Manager 应用，在「技能池」页面扫描已安装的技能。
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
