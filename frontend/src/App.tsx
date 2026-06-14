import { useState, useEffect } from "react";
import { LayoutDashboard, PackageSearch, Puzzle, Settings, Menu, X, Globe, Terminal } from "lucide-react";
import { useI18n } from "./i18n/context";
import { locales } from "./i18n/translations";
import DashboardPage from "./pages/DashboardPage";
import MarketPage from "./pages/MarketPage";
import SkillsPoolPage from "./pages/SkillsPoolPage";
import DetailPage from "./pages/DetailPage";
import SettingsPage from "./pages/SettingsPage";
import CLITipsPage from "./pages/CLITipsPage";
import type { ListedSkill, AgentInfo, SkillStats } from "./types";
import { listSkills, listAgents, getStats } from "./bridge";
import { Button } from "./components/ui/button";
import { cn } from "./lib/utils";

type Page = "dashboard" | "market" | "pool" | "detail" | "settings" | "cli-tips";

interface NavItem { key: Page; label: string; icon: typeof LayoutDashboard; }

export default function App() {
  const { t, locale, setLocale } = useI18n();
  const [page, setPage] = useState<Page>("dashboard");
  const [prevPage, setPrevPage] = useState<Page>("dashboard");
  const [skills, setSkills] = useState<ListedSkill[]>([]);
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [stats, setStats] = useState<SkillStats | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [selectedSkill, setSelectedSkill] = useState<ListedSkill | null>(null);

  const NAV_ITEMS: NavItem[] = [
    { key: "dashboard", label: t("nav.dashboard"), icon: LayoutDashboard },
    { key: "pool", label: t("nav.pool"), icon: PackageSearch },
    { key: "market", label: t("nav.market"), icon: PackageSearch },
    { key: "cli-tips", label: "CLI 教程", icon: Terminal },
    { key: "settings", label: t("nav.settings"), icon: Settings },
  ];

  const loadData = async () => {
    const [sk, ag, st] = await Promise.all([listSkills(), listAgents(), getStats()]);
    setSkills(sk);
    setAgents(ag);
    setStats(st);
  };

  useEffect(() => { loadData(); }, []);

  const navigateToDetail = (skill: ListedSkill) => {
    setSelectedSkill(skill);
    setPrevPage(page);
    setPage("detail");
  };

  const navigateBack = () => {
    setPage(prevPage);
    setSelectedSkill(null);
  };

  const navigateTo = (p: Page) => {
    setPage(p);
    setSelectedSkill(null);
  };

  const cycleLocale = () => {
    const idx = locales.findIndex((l) => l.value === locale);
    setLocale(locales[(idx + 1) % locales.length].value);
  };

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <aside className={cn("border-r bg-card flex flex-col transition-all duration-200", sidebarOpen ? "w-56" : "w-0 overflow-hidden")}>
        <div className="p-4 border-b">
          <h1 className="text-lg font-bold flex items-center gap-2">
            <Puzzle className="h-5 w-5 text-primary" />
            {t("app.title")}
          </h1>
          <p className="text-xs text-muted-foreground mt-1">{t("app.subtitle")}</p>
        </div>
        <nav className="flex-1 p-2 space-y-1">
          {NAV_ITEMS.map((item) => (
            <button key={item.key} onClick={() => navigateTo(item.key)} className={cn("w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors", page === item.key ? "bg-primary/10 text-primary font-medium" : "text-muted-foreground hover:bg-accent hover:text-accent-foreground")}>
              <item.icon className="h-4 w-4" />
              {item.label}
            </button>
          ))}
        </nav>
        <div className="p-4 border-t text-xs text-muted-foreground space-y-2">
          {stats && <span>{t("app.skills_count", { count: stats.total_skills, agents: stats.total_agents })}</span>}
        </div>
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <header className="h-12 border-b flex items-center px-4 gap-4 shrink-0">
          <Button variant="ghost" size="icon" onClick={() => setSidebarOpen(!sidebarOpen)}>
            {sidebarOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
          </Button>
          <span className="text-sm font-medium capitalize">{page}</span>
          <div className="flex-1" />
          <div className="flex items-center gap-2">
            <div className="h-4 w-px bg-border" />
            <Button
              variant="outline"
              size="sm"
              onClick={cycleLocale}
              className="gap-1.5 text-xs"
              title={locale === "zh" ? "Switch to English" : "切换到中文"}
            >
              <Globe className="h-3.5 w-3.5" />
              {locale === "zh" ? "English" : "中文"}
            </Button>
          </div>
        </header>
        <main className="flex-1 overflow-auto p-6">
          {page === "dashboard" && <DashboardPage skills={skills} agents={agents} onNavigate={navigateToDetail} />}
          {page === "market" && <MarketPage onRefresh={loadData} />}
          {page === "pool" && <SkillsPoolPage skills={skills} onSelect={navigateToDetail} onRefresh={loadData} />}
          {page === "detail" && <DetailPage skill={selectedSkill} onBack={navigateBack} />}
          {page === "settings" && <SettingsPage agents={agents} onRefresh={loadData} />}
          {page === "cli-tips" && <CLITipsPage />}
        </main>
      </div>
    </div>
  );
}
