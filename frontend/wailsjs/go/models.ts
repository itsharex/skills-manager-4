export namespace models {
	
	export class RepoSource {
	    name: string;
	    url: string;
	    type: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RepoSource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.type = source["type"];
	        this.enabled = source["enabled"];
	    }
	}
	export class LinkTarget {
	    id: string;
	    path: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LinkTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.enabled = source["enabled"];
	    }
	}
	export class MarketSource {
	    name: string;
	    url: string;
	    type: string;
	    enabled: boolean;
	    branch?: string;
	
	    static createFrom(source: any = {}) {
	        return new MarketSource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.type = source["type"];
	        this.enabled = source["enabled"];
	        this.branch = source["branch"];
	    }
	}
	export class Config {
	    repo_path: string;
	    pool_path: string;
	    install_mode: string;
	    auto_fallback: boolean;
	    default_agents: string[];
	    market_sources: MarketSource[];
	    link_targets: LinkTarget[];
	    repositories: RepoSource[];
	    cache_ttl: number;
	    github_token: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repo_path = source["repo_path"];
	        this.pool_path = source["pool_path"];
	        this.install_mode = source["install_mode"];
	        this.auto_fallback = source["auto_fallback"];
	        this.default_agents = source["default_agents"];
	        this.market_sources = this.convertValues(source["market_sources"], MarketSource);
	        this.link_targets = this.convertValues(source["link_targets"], LinkTarget);
	        this.repositories = this.convertValues(source["repositories"], RepoSource);
	        this.cache_ttl = source["cache_ttl"];
	        this.github_token = source["github_token"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class MarketSearchSkill {
	    name: string;
	    namespace: string;
	    version: string;
	    description: string;
	    source: string;
	    localPath?: string;
	    installs?: number;
	
	    static createFrom(source: any = {}) {
	        return new MarketSearchSkill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.namespace = source["namespace"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.source = source["source"];
	        this.localPath = source["localPath"];
	        this.installs = source["installs"];
	    }
	}
	export class MarketSearchResult {
	    sourceName: string;
	    sourceType: string;
	    skills: MarketSearchSkill[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new MarketSearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceName = source["sourceName"];
	        this.sourceType = source["sourceType"];
	        this.skills = this.convertValues(source["skills"], MarketSearchSkill);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class ResolvedSkill {
	    localPath: string;
	    namespace: string;
	    name: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new ResolvedSkill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.localPath = source["localPath"];
	        this.namespace = source["namespace"];
	        this.name = source["name"];
	        this.version = source["version"];
	    }
	}

}

export namespace operations {
	
	export class HealthCheckResult {
	    name: string;
	    status: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new HealthCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.message = source["message"];
	    }
	}
	export class HealthReport {
	    repo_path: string;
	    checks: HealthCheckResult[];
	    all_pass: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HealthReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repo_path = source["repo_path"];
	        this.checks = this.convertValues(source["checks"], HealthCheckResult);
	        this.all_pass = source["all_pass"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SkillStats {
	    total_skills: number;
	    total_versions: number;
	    total_namespaces: number;
	    total_agents: number;
	    installed_skills: number;
	    disk_usage_bytes: number;
	    skills_per_agent?: Record<string, number>;
	    skills_per_version?: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new SkillStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_skills = source["total_skills"];
	        this.total_versions = source["total_versions"];
	        this.total_namespaces = source["total_namespaces"];
	        this.total_agents = source["total_agents"];
	        this.installed_skills = source["installed_skills"];
	        this.disk_usage_bytes = source["disk_usage_bytes"];
	        this.skills_per_agent = source["skills_per_agent"];
	        this.skills_per_version = source["skills_per_version"];
	    }
	}

}

export namespace waillib {
	
	export class AgentInfo {
	    id: string;
	    name: string;
	    path: string;
	    skillsDir: string;
	    projectSkillsSubdir: string;
	    detected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AgentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.skillsDir = source["skillsDir"];
	        this.projectSkillsSubdir = source["projectSkillsSubdir"];
	        this.detected = source["detected"];
	    }
	}
	export class DiscoveredSkill {
	    name: string;
	    namespace: string;
	    version: string;
	    path: string;
	    agentId?: string;
	    agentName?: string;
	    alreadyInPool: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DiscoveredSkill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.namespace = source["namespace"];
	        this.version = source["version"];
	        this.path = source["path"];
	        this.agentId = source["agentId"];
	        this.agentName = source["agentName"];
	        this.alreadyInPool = source["alreadyInPool"];
	    }
	}
	export class InstallUILog {
	    skillName: string;
	    version: string;
	    path: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallUILog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillName = source["skillName"];
	        this.version = source["version"];
	        this.path = source["path"];
	        this.error = source["error"];
	    }
	}
	export class InstallUIOptions {
	    namespace: string;
	    version: string;
	    agents: string[];
	    noSync: boolean;
	
	    static createFrom(source: any = {}) {
	        return new InstallUIOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.namespace = source["namespace"];
	        this.version = source["version"];
	        this.agents = source["agents"];
	        this.noSync = source["noSync"];
	    }
	}
	export class ListedSkill {
	    name: string;
	    agentIds: string[];
	    agentNames: string[];
	    paths: string[];
	    latest: string;
	    versions: string[];
	    description: string;
	    inPool: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ListedSkill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.agentIds = source["agentIds"];
	        this.agentNames = source["agentNames"];
	        this.paths = source["paths"];
	        this.latest = source["latest"];
	        this.versions = source["versions"];
	        this.description = source["description"];
	        this.inPool = source["inPool"];
	    }
	}

}

