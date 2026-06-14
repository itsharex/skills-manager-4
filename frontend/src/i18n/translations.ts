export type Locale = "en" | "zh";

type TranslationDict = Record<string, string>;

const en: TranslationDict = {
  /* App / Sidebar */
  "app.title": "Skills Mgr",
  "app.subtitle": "v0.2 · Agent Skills Manager",
  "app.skills_count": "{count} skills · {agents} agents",

  /* Navigation */
  "nav.dashboard": "Dashboard",
  "nav.pool": "Pool",
  "nav.market": "Market",
  "nav.skills": "Skills",
  "nav.settings": "Settings",

  /* Dashboard */
  "dashboard.title": "Dashboard",
  "dashboard.subtitle": "Overview of your skill ecosystem.",
  "dashboard.total_skills": "Total Skills",
  "dashboard.installed": "Installed",
  "dashboard.namespaces": "Namespaces",
  "dashboard.agents": "Agents",
  "dashboard.disk_usage": "Disk Usage",
  "dashboard.installed_skills": "Installed Skills",
  "dashboard.no_skills": "No skills installed yet.",
  "dashboard.agents_title": "Agents",
  "dashboard.detail": "Detail",
  "dashboard.table.name": "Name",
  "dashboard.table.namespace": "Namespace",
  "dashboard.table.version": "Version",

  /* Skills Page */
  "skills.title": "Skills",
  "skills.installed_count": "{count} skill(s) installed.",
  "skills.refresh": "Refresh",
  "skills.no_skills": "No skills installed yet.",
  "skills.table.name": "Name",
  "skills.table.namespace": "Namespace",
  "skills.table.latest": "Latest",
  "skills.table.versions": "Versions",
  "skills.table.description": "Description",
  "skills.detail_btn": "Detail",

  /* Detail Page */
  "detail.title": "Skill Detail",
  "detail.no_selection": "No skill selected.",
  "detail.overview": "Overview",
  "detail.versions": "Versions",
  "detail.field.name": "Name",
  "detail.field.namespace": "Namespace",
  "detail.field.latest_version": "Latest Version",
  "detail.field.total_versions": "Total Versions",
  "detail.field.description": "Description",
  "detail.no_description": "No description.",
  "detail.no_versions": "No versions available.",
  "detail.table.version": "Version",
  "detail.table.status": "Status",
  "detail.status_latest": "Latest",

  /* Market Page */
  "market.title": "Market",
  "market.subtitle": "Search and install skills from registries or repositories.",
  "market.search_source": "Search Source",
  "market.placeholder": "GitHub URL or registry path...",
  "market.search": "Search",
  "market.results": "Search Results",
  "market.install_all": "Install All",
  "market.table.name": "Name",
  "market.table.namespace": "Namespace",
  "market.table.version": "Version",
  "market.table.path": "Path",
  "market.install_log": "Install Log",
  "market.error_prefix": "Error",

  /* Settings Page */
  "settings.title": "Settings",
  "settings.subtitle": "Manage agents and configuration.",
  "settings.agents": "Agents",
  "settings.redetect": "Re-detect",
  "settings.no_agents": "No agents detected.",
  "settings.field.id": "ID",
  "settings.detected": "Detected",
  "settings.not_detected": "Not detected",
  "settings.configuration": "Configuration",
  "settings.config_placeholder": "Additional configuration options coming soon.",
};

const zh: TranslationDict = {
  /* App / Sidebar */
  "app.title": "技能管理器",
  "app.subtitle": "v0.2 · 跨 Agent 技能管理",
  "app.skills_count": "{count} 个技能 · {agents} 个代理",

  /* Navigation */
  "nav.dashboard": "仪表盘",
  "nav.pool": "技能池",
  "nav.market": "市场",
  "nav.skills": "技能",
  "nav.settings": "设置",

  /* Dashboard */
  "dashboard.title": "仪表盘",
  "dashboard.subtitle": "技能生态系统概览。",
  "dashboard.total_skills": "技能总数",
  "dashboard.installed": "已安装",
  "dashboard.namespaces": "命名空间",
  "dashboard.agents": "代理数",
  "dashboard.disk_usage": "磁盘占用",
  "dashboard.installed_skills": "已安装技能",
  "dashboard.no_skills": "暂无已安装的技能。",
  "dashboard.agents_title": "代理",
  "dashboard.detail": "详情",
  "dashboard.table.name": "名称",
  "dashboard.table.namespace": "命名空间",
  "dashboard.table.version": "版本",

  /* Skills Page */
  "skills.title": "技能",
  "skills.installed_count": "已安装 {count} 个技能。",
  "skills.refresh": "刷新",
  "skills.no_skills": "暂无已安装的技能。",
  "skills.table.name": "名称",
  "skills.table.namespace": "命名空间",
  "skills.table.latest": "最新",
  "skills.table.versions": "版本数",
  "skills.table.description": "描述",
  "skills.detail_btn": "详情",

  /* Detail Page */
  "detail.title": "技能详情",
  "detail.no_selection": "未选择技能。",
  "detail.overview": "概览",
  "detail.versions": "版本",
  "detail.field.name": "名称",
  "detail.field.namespace": "命名空间",
  "detail.field.latest_version": "最新版本",
  "detail.field.total_versions": "版本总数",
  "detail.field.description": "描述",
  "detail.no_description": "暂无描述。",
  "detail.no_versions": "暂无可用版本。",
  "detail.table.version": "版本",
  "detail.table.status": "状态",
  "detail.status_latest": "最新",

  /* Market Page */
  "market.title": "市场",
  "market.subtitle": "从注册表或仓库搜索并安装技能。",
  "market.search_source": "搜索来源",
  "market.placeholder": "GitHub URL 或注册表路径...",
  "market.search": "搜索",
  "market.results": "搜索结果",
  "market.install_all": "全部安装",
  "market.table.name": "名称",
  "market.table.namespace": "命名空间",
  "market.table.version": "版本",
  "market.table.path": "路径",
  "market.install_log": "安装日志",
  "market.error_prefix": "错误",

  /* Settings Page */
  "settings.title": "设置",
  "settings.subtitle": "管理 Agent 和配置。",
  "settings.agents": "代理",
  "settings.redetect": "重新检测",
  "settings.no_agents": "未检测到 Agent。",
  "settings.field.id": "ID",
  "settings.detected": "已检测",
  "settings.not_detected": "未检测",
  "settings.configuration": "配置",
  "settings.config_placeholder": "更多配置选项即将推出。",
};

const dictionaries: Record<Locale, TranslationDict> = { en, zh };

export function translate(key: string, locale: Locale, params?: Record<string, string | number>): string {
  const dict = dictionaries[locale];
  let text = dict[key] ?? key;
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      text = text.replace(`{${k}}`, String(v));
    }
  }
  return text;
}

export const locales: { value: Locale; label: Record<Locale, string> }[] = [
  { value: "en", label: { en: "English", zh: "英文" } },
  { value: "zh", label: { en: "Chinese", zh: "中文" } },
];