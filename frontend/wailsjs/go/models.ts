export namespace api {
	
	export class InstallLink {
	    agent_id: string;
	    path: string;
	    success: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agent_id = source["agent_id"];
	        this.path = source["path"];
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class InstallRequest {
	    source: string;
	    sub_path?: string;
	    version?: string;
	    ref?: string;
	    agents: string[];
	    scope: string;
	    project_dir?: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.sub_path = source["sub_path"];
	        this.version = source["version"];
	        this.ref = source["ref"];
	        this.agents = source["agents"];
	        this.scope = source["scope"];
	        this.project_dir = source["project_dir"];
	    }
	}
	export class InstallResult {
	    skill_name: string;
	    version: string;
	    source: models.Source;
	    agents: Record<string, InstallLink>;
	
	    static createFrom(source: any = {}) {
	        return new InstallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skill_name = source["skill_name"];
	        this.version = source["version"];
	        this.source = this.convertValues(source["source"], models.Source);
	        this.agents = this.convertValues(source["agents"], InstallLink, true);
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

}

export namespace models {
	
	export class Agent {
	    name: string;
	    skill_location: string;
	    global_location: string;
	    installed: boolean;
	    detected: boolean;
	    supports_project: boolean;
	    global_directory_key: string;
	    project_directory_key: string;
	
	    static createFrom(source: any = {}) {
	        return new Agent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.skill_location = source["skill_location"];
	        this.global_location = source["global_location"];
	        this.installed = source["installed"];
	        this.detected = source["detected"];
	        this.supports_project = source["supports_project"];
	        this.global_directory_key = source["global_directory_key"];
	        this.project_directory_key = source["project_directory_key"];
	    }
	}
	export class AgentInstall {
	    agentId: string;
	    version: string;
	    // Go type: time
	    installedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new AgentInstall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agentId = source["agentId"];
	        this.version = source["version"];
	        this.installedAt = this.convertValues(source["installedAt"], null);
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
	export class AgentSkillStatus {
	    agentId: string;
	    agentName: string;
	    path: string;
	    version?: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentSkillStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agentId = source["agentId"];
	        this.agentName = source["agentName"];
	        this.path = source["path"];
	        this.version = source["version"];
	    }
	}
	export class SkillInstall {
	    skillName: string;
	    version: string;
	    isLatest: boolean;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillInstall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillName = source["skillName"];
	        this.version = source["version"];
	        this.isLatest = source["isLatest"];
	        this.status = source["status"];
	    }
	}
	export class AgentStats {
	    id: string;
	    name: string;
	    skillCount: number;
	    orphanedCount: number;
	    totalSizeBytes: number;
	    installedSkills: SkillInstall[];
	
	    static createFrom(source: any = {}) {
	        return new AgentStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.skillCount = source["skillCount"];
	        this.orphanedCount = source["orphanedCount"];
	        this.totalSizeBytes = source["totalSizeBytes"];
	        this.installedSkills = this.convertValues(source["installedSkills"], SkillInstall);
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
	export class BatchCleanCriteria {
	    unused: boolean;
	    olderThanDays: number;
	    namePattern: string;
	    sourceType: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchCleanCriteria(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.unused = source["unused"];
	        this.olderThanDays = source["olderThanDays"];
	        this.namePattern = source["namePattern"];
	        this.sourceType = source["sourceType"];
	    }
	}
	export class CleanItemResult {
	    skillName: string;
	    version?: string;
	    action: string;
	    success: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new CleanItemResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillName = source["skillName"];
	        this.version = source["version"];
	        this.action = source["action"];
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class BatchCleanResult {
	    total: number;
	    succeeded: number;
	    failed: number;
	    results: CleanItemResult[];
	    dryRun: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BatchCleanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.succeeded = source["succeeded"];
	        this.failed = source["failed"];
	        this.results = this.convertValues(source["results"], CleanItemResult);
	        this.dryRun = source["dryRun"];
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
	export class BatchSyncItemResult {
	    skillName: string;
	    agentId: string;
	    success: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchSyncItemResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillName = source["skillName"];
	        this.agentId = source["agentId"];
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class BatchSyncRequest {
	    skillNames: string[];
	    agentIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new BatchSyncRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillNames = source["skillNames"];
	        this.agentIds = source["agentIds"];
	    }
	}
	export class BatchSyncResult {
	    total: number;
	    succeeded: number;
	    failed: number;
	    results: BatchSyncItemResult[];
	
	    static createFrom(source: any = {}) {
	        return new BatchSyncResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.succeeded = source["succeeded"];
	        this.failed = source["failed"];
	        this.results = this.convertValues(source["results"], BatchSyncItemResult);
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
	export class ClawHubSkill {
	    owner: string;
	    slug: string;
	    name: string;
	    description: string;
	    version?: string;
	    author?: string;
	    tags?: string[];
	    downloads?: number;
	    stars?: number;
	    updatedAt?: string;
	
	    static createFrom(source: any = {}) {
	        return new ClawHubSkill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.owner = source["owner"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.tags = source["tags"];
	        this.downloads = source["downloads"];
	        this.stars = source["stars"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	
	export class CleanResult {
	    totalProcessed: number;
	    succeeded: number;
	    failed: number;
	    items: CleanItemResult[];
	    errors?: string[];
	
	    static createFrom(source: any = {}) {
	        return new CleanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalProcessed = source["totalProcessed"];
	        this.succeeded = source["succeeded"];
	        this.failed = source["failed"];
	        this.items = this.convertValues(source["items"], CleanItemResult);
	        this.errors = source["errors"];
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
	export class SkillMarketConfig {
	    url: string;
	    cacheEnabled: boolean;
	    cacheExpiryHours: number;
	    lastUpdated?: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillMarketConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.cacheEnabled = source["cacheEnabled"];
	        this.cacheExpiryHours = source["cacheExpiryHours"];
	        this.lastUpdated = source["lastUpdated"];
	    }
	}
	export class SkillspoolConfig {
	    root: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillspoolConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	    }
	}
	export class Config {
	    skillspool: SkillspoolConfig;
	    skillMarket: SkillMarketConfig;
	    agents: Record<string, Agent>;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillspool = this.convertValues(source["skillspool"], SkillspoolConfig);
	        this.skillMarket = this.convertValues(source["skillMarket"], SkillMarketConfig);
	        this.agents = this.convertValues(source["agents"], Agent, true);
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
	export class FileIssue {
	    skillName: string;
	    path: string;
	    missing: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillName = source["skillName"];
	        this.path = source["path"];
	        this.missing = source["missing"];
	    }
	}
	export class GlobalSkillWithAgents {
	    name: string;
	    description: string;
	    installedAgents: AgentSkillStatus[];
	
	    static createFrom(source: any = {}) {
	        return new GlobalSkillWithAgents(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.installedAgents = this.convertValues(source["installedAgents"], AgentSkillStatus);
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
	export class HealthIssue {
	    type: string;
	    severity: string;
	    skillName?: string;
	    agentId?: string;
	    path: string;
	    message: string;
	    remediation?: string;
	
	    static createFrom(source: any = {}) {
	        return new HealthIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.severity = source["severity"];
	        this.skillName = source["skillName"];
	        this.agentId = source["agentId"];
	        this.path = source["path"];
	        this.message = source["message"];
	        this.remediation = source["remediation"];
	    }
	}
	export class VersionIssue {
	    skillName: string;
	    version: string;
	    issue: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillName = source["skillName"];
	        this.version = source["version"];
	        this.issue = source["issue"];
	    }
	}
	export class SymlinkIssue {
	    agentId: string;
	    path: string;
	    target: string;
	    targetExists: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SymlinkIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agentId = source["agentId"];
	        this.path = source["path"];
	        this.target = source["target"];
	        this.targetExists = source["targetExists"];
	    }
	}
	export class HealthSummary {
	    totalSkills: number;
	    totalAgents: number;
	    brokenSymlinks: number;
	    missingFiles: number;
	    unreachableSkills: number;
	
	    static createFrom(source: any = {}) {
	        return new HealthSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSkills = source["totalSkills"];
	        this.totalAgents = source["totalAgents"];
	        this.brokenSymlinks = source["brokenSymlinks"];
	        this.missingFiles = source["missingFiles"];
	        this.unreachableSkills = source["unreachableSkills"];
	    }
	}
	export class HealthReport {
	    // Go type: time
	    generatedAt: any;
	    status: string;
	    summary: HealthSummary;
	    issues: HealthIssue[];
	    symlinks?: SymlinkIssue[];
	    files?: FileIssue[];
	    versions?: VersionIssue[];
	
	    static createFrom(source: any = {}) {
	        return new HealthReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generatedAt = this.convertValues(source["generatedAt"], null);
	        this.status = source["status"];
	        this.summary = this.convertValues(source["summary"], HealthSummary);
	        this.issues = this.convertValues(source["issues"], HealthIssue);
	        this.symlinks = this.convertValues(source["symlinks"], SymlinkIssue);
	        this.files = this.convertValues(source["files"], FileIssue);
	        this.versions = this.convertValues(source["versions"], VersionIssue);
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
	
	export class MarketSkill {
	    name: string;
	    description: string;
	    author?: string;
	    version?: string;
	    category?: string;
	    tags?: string[];
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new MarketSkill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.author = source["author"];
	        this.version = source["version"];
	        this.category = source["category"];
	        this.tags = source["tags"];
	        this.source = source["source"];
	    }
	}
	export class MigrateResult {
	    success: boolean;
	    libraryPath: string;
	    symlinkCreated: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new MigrateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.libraryPath = source["libraryPath"];
	        this.symlinkCreated = source["symlinkCreated"];
	        this.error = source["error"];
	    }
	}
	export class ProjectSkill {
	    name: string;
	    description: string;
	    path: string;
	    isSymlink: boolean;
	    symlinkTarget?: string;
	    inLibrary: boolean;
	    version?: string;
	    tags?: string[];
	    sizeBytes?: number;
	
	    static createFrom(source: any = {}) {
	        return new ProjectSkill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.path = source["path"];
	        this.isSymlink = source["isSymlink"];
	        this.symlinkTarget = source["symlinkTarget"];
	        this.inLibrary = source["inLibrary"];
	        this.version = source["version"];
	        this.tags = source["tags"];
	        this.sizeBytes = source["sizeBytes"];
	    }
	}
	export class RuntimeStatus {
	    nodeInstalled: boolean;
	    nodeVersion?: string;
	    nodePath?: string;
	    clawhubInstalled: boolean;
	    clawhubVersion?: string;
	    clawhubPath?: string;
	    hasNpm: boolean;
	    message?: string;
	    registryReachable?: boolean;
	    registryName?: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeInstalled = source["nodeInstalled"];
	        this.nodeVersion = source["nodeVersion"];
	        this.nodePath = source["nodePath"];
	        this.clawhubInstalled = source["clawhubInstalled"];
	        this.clawhubVersion = source["clawhubVersion"];
	        this.clawhubPath = source["clawhubPath"];
	        this.hasNpm = source["hasNpm"];
	        this.message = source["message"];
	        this.registryReachable = source["registryReachable"];
	        this.registryName = source["registryName"];
	    }
	}
	export class ScanMarketResult {
	    totalSkills: number;
	    categories: string[];
	    skills: MarketSkill[];
	
	    static createFrom(source: any = {}) {
	        return new ScanMarketResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSkills = source["totalSkills"];
	        this.categories = source["categories"];
	        this.skills = this.convertValues(source["skills"], MarketSkill);
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
	export class SkillVersion {
	    version: string;
	    installed_at: string;
	    path: string;
	    agents: string[];
	
	    static createFrom(source: any = {}) {
	        return new SkillVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.installed_at = source["installed_at"];
	        this.path = source["path"];
	        this.agents = source["agents"];
	    }
	}
	export class Source {
	    type: string;
	    url?: string;
	    ref?: string;
	    sub_path?: string;
	    command?: string;
	    path?: string;
	
	    static createFrom(source: any = {}) {
	        return new Source(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.url = source["url"];
	        this.ref = source["ref"];
	        this.sub_path = source["sub_path"];
	        this.command = source["command"];
	        this.path = source["path"];
	    }
	}
	export class Skill {
	    name: string;
	    description: string;
	    source: Source;
	    versions: Record<string, SkillVersion>;
	    latest_version: string;
	    user_tags?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Skill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.source = this.convertValues(source["source"], Source);
	        this.versions = this.convertValues(source["versions"], SkillVersion, true);
	        this.latest_version = source["latest_version"];
	        this.user_tags = source["user_tags"];
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
	export class SkillActivity {
	    skillName: string;
	    // Go type: time
	    lastActivity: any;
	
	    static createFrom(source: any = {}) {
	        return new SkillActivity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillName = source["skillName"];
	        this.lastActivity = this.convertValues(source["lastActivity"], null);
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
	export class SkillCount {
	    skillName: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new SkillCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillName = source["skillName"];
	        this.count = source["count"];
	    }
	}
	
	
	export class SkillStats {
	    name: string;
	    versionCount: number;
	    currentVersion: string;
	    sizeBytes: number;
	    installedBy: AgentInstall[];
	
	    static createFrom(source: any = {}) {
	        return new SkillStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.versionCount = source["versionCount"];
	        this.currentVersion = source["currentVersion"];
	        this.sizeBytes = source["sizeBytes"];
	        this.installedBy = this.convertValues(source["installedBy"], AgentInstall);
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
	
	export class SkillWithStatus {
	    name: string;
	    description: string;
	    tags?: string[];
	    latestVersion: string;
	    installStatus: string;
	    installedAgents: string[];
	    totalAgents: number;
	    source: Source;
	    sizeBytes?: number;
	
	    static createFrom(source: any = {}) {
	        return new SkillWithStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.latestVersion = source["latestVersion"];
	        this.installStatus = source["installStatus"];
	        this.installedAgents = source["installedAgents"];
	        this.totalAgents = source["totalAgents"];
	        this.source = this.convertValues(source["source"], Source);
	        this.sizeBytes = source["sizeBytes"];
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
	
	export class SkillspoolMigrationResult {
	    success: boolean;
	    old_root: string;
	    new_root: string;
	    moved_files: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillspoolMigrationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.old_root = source["old_root"];
	        this.new_root = source["new_root"];
	        this.moved_files = source["moved_files"];
	        this.message = source["message"];
	    }
	}
	
	
	export class TagUsage {
	    tag: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new TagUsage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag = source["tag"];
	        this.count = source["count"];
	    }
	}
	export class UsageDashboard {
	    totalSkills: number;
	    totalInstallations: number;
	    totalSizeBytes: number;
	    averagePerAgent: number;
	    mostPopular: SkillCount[];
	    leastUsed: SkillCount[];
	    recentlyActive: SkillActivity[];
	
	    static createFrom(source: any = {}) {
	        return new UsageDashboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSkills = source["totalSkills"];
	        this.totalInstallations = source["totalInstallations"];
	        this.totalSizeBytes = source["totalSizeBytes"];
	        this.averagePerAgent = source["averagePerAgent"];
	        this.mostPopular = this.convertValues(source["mostPopular"], SkillCount);
	        this.leastUsed = this.convertValues(source["leastUsed"], SkillCount);
	        this.recentlyActive = this.convertValues(source["recentlyActive"], SkillActivity);
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
	export class VersionInfo {
	    version: string;
	    // Go type: time
	    installed: any;
	    sizeBytes: number;
	    source: string;
	    isLatest: boolean;
	    agentCount: number;
	
	    static createFrom(source: any = {}) {
	        return new VersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.installed = this.convertValues(source["installed"], null);
	        this.sizeBytes = source["sizeBytes"];
	        this.source = source["source"];
	        this.isLatest = source["isLatest"];
	        this.agentCount = source["agentCount"];
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

}

