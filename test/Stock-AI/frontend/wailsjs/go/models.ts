export namespace data {
	
	export class SentimentComponent {
	    name: string;
	    nameCn: string;
	    value: number;
	    weight: number;
	    data: string;
	
	    static createFrom(source: any = {}) {
	        return new SentimentComponent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.nameCn = source["nameCn"];
	        this.value = source["value"];
	        this.weight = source["weight"];
	        this.data = source["data"];
	    }
	}
	export class MarketSentiment {
	    value: number;
	    level: string;
	    levelCn: string;
	    description: string;
	    updateTime: string;
	    components: SentimentComponent[];
	
	    static createFrom(source: any = {}) {
	        return new MarketSentiment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.level = source["level"];
	        this.levelCn = source["levelCn"];
	        this.description = source["description"];
	        this.updateTime = source["updateTime"];
	        this.components = this.convertValues(source["components"], SentimentComponent);
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

export namespace gorm {
	
	export class DeletedAt {
	    // Go type: time
	    Time: any;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeletedAt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = this.convertValues(source["Time"], null);
	        this.Valid = source["Valid"];
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

export namespace main {
	
	export class AIChatSession {
	    sessionId: string;
	    messages: models.AIMessage[];
	    createdAt: string;
	    lastMessage: string;
	    messageCount: number;
	
	    static createFrom(source: any = {}) {
	        return new AIChatSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.messages = this.convertValues(source["messages"], models.AIMessage);
	        this.createdAt = source["createdAt"];
	        this.lastMessage = source["lastMessage"];
	        this.messageCount = source["messageCount"];
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
	export class CachedGlobalMarketData {
	    globalIndices: models.GlobalIndex[];
	    news: models.NewsItem[];
	    sentiment?: data.MarketSentiment;
	    cacheTime: string;
	    hasCache: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CachedGlobalMarketData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.globalIndices = this.convertValues(source["globalIndices"], models.GlobalIndex);
	        this.news = this.convertValues(source["news"], models.NewsItem);
	        this.sentiment = this.convertValues(source["sentiment"], data.MarketSentiment);
	        this.cacheTime = source["cacheTime"];
	        this.hasCache = source["hasCache"];
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
	export class CachedMarketData {
	    marketIndex: models.MarketIndex[];
	    industryRank: models.IndustryRank[];
	    moneyFlow: models.MoneyFlow[];
	    newsList: models.NewsItem[];
	    longTigerRank: models.LongTigerItem[];
	    hotTopics: models.HotTopic[];
	    cacheTime: string;
	    hasCache: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CachedMarketData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.marketIndex = this.convertValues(source["marketIndex"], models.MarketIndex);
	        this.industryRank = this.convertValues(source["industryRank"], models.IndustryRank);
	        this.moneyFlow = this.convertValues(source["moneyFlow"], models.MoneyFlow);
	        this.newsList = this.convertValues(source["newsList"], models.NewsItem);
	        this.longTigerRank = this.convertValues(source["longTigerRank"], models.LongTigerItem);
	        this.hotTopics = this.convertValues(source["hotTopics"], models.HotTopic);
	        this.cacheTime = source["cacheTime"];
	        this.hasCache = source["hasCache"];
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
	export class DataCleanupInfo {
	    cacheStats: Record<string, any>;
	    rateLimiterStats: Record<string, any>;
	    aiDataInfo: Record<string, any>;
	    cleanupConfig: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new DataCleanupInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cacheStats = source["cacheStats"];
	        this.rateLimiterStats = source["rateLimiterStats"];
	        this.aiDataInfo = source["aiDataInfo"];
	        this.cleanupConfig = source["cleanupConfig"];
	    }
	}
	export class TradingTimeInfo {
	    isTradingTime: boolean;
	    isPreMarketTime: boolean;
	    refreshInterval: number;
	
	    static createFrom(source: any = {}) {
	        return new TradingTimeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isTradingTime = source["isTradingTime"];
	        this.isPreMarketTime = source["isPreMarketTime"];
	        this.refreshInterval = source["refreshInterval"];
	    }
	}

}

export namespace models {
	
	export class AIAnalysisResult {
	    id: number;
	    stockCode: string;
	    stockName: string;
	    analysis: string;
	    suggestion: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new AIAnalysisResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.analysis = source["analysis"];
	        this.suggestion = source["suggestion"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class AIChatRequest {
	    message: string;
	    sessionId: string;
	    stockCode?: string;
	    fundCode?: string;
	
	    static createFrom(source: any = {}) {
	        return new AIChatRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.sessionId = source["sessionId"];
	        this.stockCode = source["stockCode"];
	        this.fundCode = source["fundCode"];
	    }
	}
	export class AIChatResponse {
	    content: string;
	    sessionId: string;
	    done: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AIChatResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.sessionId = source["sessionId"];
	        this.done = source["done"];
	    }
	}
	export class AIMessage {
	    id: number;
	    sessionId: string;
	    role: string;
	    content: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new AIMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sessionId = source["sessionId"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class AlertNotification {
	    id: number;
	    stockCode: string;
	    stockName: string;
	    alertType: string;
	    targetValue: number;
	    currentPrice: number;
	    currentChange: number;
	    message: string;
	    time: string;
	    assetType: string;
	
	    static createFrom(source: any = {}) {
	        return new AlertNotification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.alertType = source["alertType"];
	        this.targetValue = source["targetValue"];
	        this.currentPrice = source["currentPrice"];
	        this.currentChange = source["currentChange"];
	        this.message = source["message"];
	        this.time = source["time"];
	        this.assetType = source["assetType"];
	    }
	}
	export class Config {
	    id: number;
	    refreshInterval: number;
	    proxyUrl: string;
	    proxyPoolEnabled: boolean;
	    proxyProvider: string;
	    proxyApiUrl: string;
	    proxyApiKey: string;
	    proxyApiSecret: string;
	    proxyRegion: string;
	    proxyPoolList: string;
	    proxyPoolProtocol: string;
	    proxyPoolTTL: number;
	    proxyPoolSize: number;
	    theme: string;
	    customPrimary: string;
	    alertPushEnabled: boolean;
	    wecomWebhook: string;
	    dingtalkWebhook: string;
	    emailPushEnabled: boolean;
	    emailSmtp: string;
	    emailPort: number;
	    emailUser: string;
	    emailPassword: string;
	    emailTo: string;
	    aiEnabled: boolean;
	    aiModel: string;
	    aiApiKey: string;
	    aiApiUrl: string;
	    browserPath: string;
	    paidApiEnabled: boolean;
	    paidApiProvider: string;
	    paidApiKey: string;
	    paidApiSecret: string;
	    paidApiUrl: string;
	    tushareToken: string;
	    tushareEnabled: boolean;
	    akshareEnabled: boolean;
	    dataSourcePriority: string;
	    activePersona: string;
	    skipUpdateVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.refreshInterval = source["refreshInterval"];
	        this.proxyUrl = source["proxyUrl"];
	        this.proxyPoolEnabled = source["proxyPoolEnabled"];
	        this.proxyProvider = source["proxyProvider"];
	        this.proxyApiUrl = source["proxyApiUrl"];
	        this.proxyApiKey = source["proxyApiKey"];
	        this.proxyApiSecret = source["proxyApiSecret"];
	        this.proxyRegion = source["proxyRegion"];
	        this.proxyPoolList = source["proxyPoolList"];
	        this.proxyPoolProtocol = source["proxyPoolProtocol"];
	        this.proxyPoolTTL = source["proxyPoolTTL"];
	        this.proxyPoolSize = source["proxyPoolSize"];
	        this.theme = source["theme"];
	        this.customPrimary = source["customPrimary"];
	        this.alertPushEnabled = source["alertPushEnabled"];
	        this.wecomWebhook = source["wecomWebhook"];
	        this.dingtalkWebhook = source["dingtalkWebhook"];
	        this.emailPushEnabled = source["emailPushEnabled"];
	        this.emailSmtp = source["emailSmtp"];
	        this.emailPort = source["emailPort"];
	        this.emailUser = source["emailUser"];
	        this.emailPassword = source["emailPassword"];
	        this.emailTo = source["emailTo"];
	        this.aiEnabled = source["aiEnabled"];
	        this.aiModel = source["aiModel"];
	        this.aiApiKey = source["aiApiKey"];
	        this.aiApiUrl = source["aiApiUrl"];
	        this.browserPath = source["browserPath"];
	        this.paidApiEnabled = source["paidApiEnabled"];
	        this.paidApiProvider = source["paidApiProvider"];
	        this.paidApiKey = source["paidApiKey"];
	        this.paidApiSecret = source["paidApiSecret"];
	        this.paidApiUrl = source["paidApiUrl"];
	        this.tushareToken = source["tushareToken"];
	        this.tushareEnabled = source["tushareEnabled"];
	        this.akshareEnabled = source["akshareEnabled"];
	        this.dataSourcePriority = source["dataSourcePriority"];
	        this.activePersona = source["activePersona"];
	        this.skipUpdateVersion = source["skipUpdateVersion"];
	    }
	}
	export class ProxyStatus {
	    enabled: boolean;
	    poolEnabled: boolean;
	    provider: string;
	    activeProxies: number;
	    expiresAt: string;
	    expiresInSeconds: number;
	    lastFetch: string;
	    lastError: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.poolEnabled = source["poolEnabled"];
	        this.provider = source["provider"];
	        this.activeProxies = source["activeProxies"];
	        this.expiresAt = source["expiresAt"];
	        this.expiresInSeconds = source["expiresInSeconds"];
	        this.lastFetch = source["lastFetch"];
	        this.lastError = source["lastError"];
	    }
	}
	export class DataSourceStatus {
	    key: string;
	    name: string;
	    domain: string;
	    latency: number;
	    latencyLabel: string;
	    lastChecked: string;
	    lastSuccess: string;
	    status: string;
	    failCount: number;
	
	    static createFrom(source: any = {}) {
	        return new DataSourceStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.domain = source["domain"];
	        this.latency = source["latency"];
	        this.latencyLabel = source["latencyLabel"];
	        this.lastChecked = source["lastChecked"];
	        this.lastSuccess = source["lastSuccess"];
	        this.status = source["status"];
	        this.failCount = source["failCount"];
	    }
	}
	export class DataPipelineStatus {
	    marketSources: DataSourceStatus[];
	    financial: Record<string, any>;
	    proxy: ProxyStatus;
	    generatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new DataPipelineStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.marketSources = this.convertValues(source["marketSources"], DataSourceStatus);
	        this.financial = source["financial"];
	        this.proxy = this.convertValues(source["proxy"], ProxyStatus);
	        this.generatedAt = source["generatedAt"];
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
	
	export class ForexRate {
	    pair: string;
	    name: string;
	    rate: number;
	    change: number;
	    changePercent: number;
	    high: number;
	    low: number;
	    updateTime: string;
	
	    static createFrom(source: any = {}) {
	        return new ForexRate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pair = source["pair"];
	        this.name = source["name"];
	        this.rate = source["rate"];
	        this.change = source["change"];
	        this.changePercent = source["changePercent"];
	        this.high = source["high"];
	        this.low = source["low"];
	        this.updateTime = source["updateTime"];
	    }
	}
	export class Fund {
	    id: number;
	    code: string;
	    name: string;
	    type: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Fund(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class FundAlert {
	    id: number;
	    fundCode: string;
	    fundName: string;
	    alertType: string;
	    targetValue: number;
	    condition: string;
	    enabled: boolean;
	    triggered: boolean;
	    // Go type: time
	    triggeredAt?: any;
	    triggeredNav: number;
	    triggeredChange: number;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new FundAlert(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fundCode = source["fundCode"];
	        this.fundName = source["fundName"];
	        this.alertType = source["alertType"];
	        this.targetValue = source["targetValue"];
	        this.condition = source["condition"];
	        this.enabled = source["enabled"];
	        this.triggered = source["triggered"];
	        this.triggeredAt = this.convertValues(source["triggeredAt"], null);
	        this.triggeredNav = source["triggeredNav"];
	        this.triggeredChange = source["triggeredChange"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class FundDetail {
	    code: string;
	    name: string;
	    type: string;
	    riskLevel: string;
	    manager: string;
	    company: string;
	    scale: number;
	    scaleDate: string;
	    inceptionDate: string;
	    navDate: string;
	    nav: number;
	    estimate: number;
	    oneDayReturn: number;
	    oneYearReturn: number;
	    threeYearReturn: number;
	    thisYearReturn: number;
	    sinceStartReturn: number;
	    sharpRatio: number;
	    maxDrawdown: number;
	
	    static createFrom(source: any = {}) {
	        return new FundDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.riskLevel = source["riskLevel"];
	        this.manager = source["manager"];
	        this.company = source["company"];
	        this.scale = source["scale"];
	        this.scaleDate = source["scaleDate"];
	        this.inceptionDate = source["inceptionDate"];
	        this.navDate = source["navDate"];
	        this.nav = source["nav"];
	        this.estimate = source["estimate"];
	        this.oneDayReturn = source["oneDayReturn"];
	        this.oneYearReturn = source["oneYearReturn"];
	        this.threeYearReturn = source["threeYearReturn"];
	        this.thisYearReturn = source["thisYearReturn"];
	        this.sinceStartReturn = source["sinceStartReturn"];
	        this.sharpRatio = source["sharpRatio"];
	        this.maxDrawdown = source["maxDrawdown"];
	    }
	}
	export class FundHolding {
	    code: string;
	    name: string;
	    ratio: number;
	    industry: string;
	    trend: string;
	    change: number;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new FundHolding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.ratio = source["ratio"];
	        this.industry = source["industry"];
	        this.trend = source["trend"];
	        this.change = source["change"];
	        this.type = source["type"];
	    }
	}
	export class FundNotice {
	    id: string;
	    title: string;
	    date: string;
	    category: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new FundNotice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.date = source["date"];
	        this.category = source["category"];
	        this.url = source["url"];
	    }
	}
	export class FundPerformancePoint {
	    date: string;
	    nav: number;
	    accNav: number;
	    changePercent: number;
	
	    static createFrom(source: any = {}) {
	        return new FundPerformancePoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.nav = source["nav"];
	        this.accNav = source["accNav"];
	        this.changePercent = source["changePercent"];
	    }
	}
	export class FundPrice {
	    code: string;
	    name: string;
	    nav: number;
	    estimate: number;
	    changePercent: number;
	    updateTime: string;
	
	    static createFrom(source: any = {}) {
	        return new FundPrice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.nav = source["nav"];
	        this.estimate = source["estimate"];
	        this.changePercent = source["changePercent"];
	        this.updateTime = source["updateTime"];
	    }
	}
	export class FundOverview {
	    price?: FundPrice;
	    detail?: FundDetail;
	    history: FundPerformancePoint[];
	    stockHoldings: FundHolding[];
	    bondHoldings: FundHolding[];
	    notices: FundNotice[];
	
	    static createFrom(source: any = {}) {
	        return new FundOverview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.price = this.convertValues(source["price"], FundPrice);
	        this.detail = this.convertValues(source["detail"], FundDetail);
	        this.history = this.convertValues(source["history"], FundPerformancePoint);
	        this.stockHoldings = this.convertValues(source["stockHoldings"], FundHolding);
	        this.bondHoldings = this.convertValues(source["bondHoldings"], FundHolding);
	        this.notices = this.convertValues(source["notices"], FundNotice);
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
	
	export class FundPosition {
	    id: number;
	    fundCode: string;
	    fundName: string;
	    buyNav: number;
	    buyDate: string;
	    share: number;
	    costNav: number;
	    targetNav: number;
	    stopLossNav: number;
	    notes: string;
	    status: string;
	    redeemNav: number;
	    redeemDate: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new FundPosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fundCode = source["fundCode"];
	        this.fundName = source["fundName"];
	        this.buyNav = source["buyNav"];
	        this.buyDate = source["buyDate"];
	        this.share = source["share"];
	        this.costNav = source["costNav"];
	        this.targetNav = source["targetNav"];
	        this.stopLossNav = source["stopLossNav"];
	        this.notes = source["notes"];
	        this.status = source["status"];
	        this.redeemNav = source["redeemNav"];
	        this.redeemDate = source["redeemDate"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	
	export class Futures {
	    id: number;
	    code: string;
	    name: string;
	    exchange: string;
	    product: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Futures(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.exchange = source["exchange"];
	        this.product = source["product"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class FuturesPrice {
	    code: string;
	    name: string;
	    price: number;
	    change: number;
	    changePercent: number;
	    open: number;
	    high: number;
	    low: number;
	    preClose: number;
	    preSettle: number;
	    settle: number;
	    volume: number;
	    amount: number;
	    openInterest: number;
	    updateTime: string;
	    exchange: string;
	
	    static createFrom(source: any = {}) {
	        return new FuturesPrice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.price = source["price"];
	        this.change = source["change"];
	        this.changePercent = source["changePercent"];
	        this.open = source["open"];
	        this.high = source["high"];
	        this.low = source["low"];
	        this.preClose = source["preClose"];
	        this.preSettle = source["preSettle"];
	        this.settle = source["settle"];
	        this.volume = source["volume"];
	        this.amount = source["amount"];
	        this.openInterest = source["openInterest"];
	        this.updateTime = source["updateTime"];
	        this.exchange = source["exchange"];
	    }
	}
	export class FuturesProduct {
	    code: string;
	    name: string;
	    exchange: string;
	    unit: string;
	    margin: string;
	
	    static createFrom(source: any = {}) {
	        return new FuturesProduct(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.exchange = source["exchange"];
	        this.unit = source["unit"];
	        this.margin = source["margin"];
	    }
	}
	export class GlobalIndex {
	    code: string;
	    name: string;
	    nameCn: string;
	    price: number;
	    change: number;
	    changePercent: number;
	    open: number;
	    high: number;
	    low: number;
	    preClose: number;
	    updateTime: string;
	    region: string;
	    country: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new GlobalIndex(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.nameCn = source["nameCn"];
	        this.price = source["price"];
	        this.change = source["change"];
	        this.changePercent = source["changePercent"];
	        this.open = source["open"];
	        this.high = source["high"];
	        this.low = source["low"];
	        this.preClose = source["preClose"];
	        this.updateTime = source["updateTime"];
	        this.region = source["region"];
	        this.country = source["country"];
	        this.status = source["status"];
	    }
	}
	export class HKStock {
	    id: number;
	    code: string;
	    name: string;
	    nameCn: string;
	    lot: number;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new HKStock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.nameCn = source["nameCn"];
	        this.lot = source["lot"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class HotTopic {
	    rank: number;
	    title: string;
	    desc: string;
	    readCount: number;
	    postCount: number;
	
	    static createFrom(source: any = {}) {
	        return new HotTopic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rank = source["rank"];
	        this.title = source["title"];
	        this.desc = source["desc"];
	        this.readCount = source["readCount"];
	        this.postCount = source["postCount"];
	    }
	}
	export class IndustryRank {
	    name: string;
	    changePercent: number;
	    leadStock: string;
	
	    static createFrom(source: any = {}) {
	        return new IndustryRank(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.changePercent = source["changePercent"];
	        this.leadStock = source["leadStock"];
	    }
	}
	export class KLineData {
	    date: string;
	    open: number;
	    high: number;
	    low: number;
	    close: number;
	    volume: number;
	    code: string;
	
	    static createFrom(source: any = {}) {
	        return new KLineData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.open = source["open"];
	        this.high = source["high"];
	        this.low = source["low"];
	        this.close = source["close"];
	        this.volume = source["volume"];
	        this.code = source["code"];
	    }
	}
	export class LongTigerItem {
	    rank: number;
	    code: string;
	    name: string;
	    changePercent: number;
	    buyAmount: string;
	    sellAmount: string;
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new LongTigerItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rank = source["rank"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.changePercent = source["changePercent"];
	        this.buyAmount = source["buyAmount"];
	        this.sellAmount = source["sellAmount"];
	        this.date = source["date"];
	    }
	}
	export class MarketIndex {
	    code: string;
	    name: string;
	    price: number;
	    change: number;
	    changePercent: number;
	
	    static createFrom(source: any = {}) {
	        return new MarketIndex(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.price = source["price"];
	        this.change = source["change"];
	        this.changePercent = source["changePercent"];
	    }
	}
	export class MinuteData {
	    time: string;
	    price: number;
	    volume: number;
	    changePercent: number;
	
	    static createFrom(source: any = {}) {
	        return new MinuteData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.price = source["price"];
	        this.volume = source["volume"];
	        this.changePercent = source["changePercent"];
	    }
	}
	export class MoneyFlow {
	    code: string;
	    name: string;
	    mainFlow: number;
	    superFlow: number;
	
	    static createFrom(source: any = {}) {
	        return new MoneyFlow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.mainFlow = source["mainFlow"];
	        this.superFlow = source["superFlow"];
	    }
	}
	export class NewsItem {
	    id: number;
	    title: string;
	    content: string;
	    time: string;
	    source: string;
	    importance: string;
	
	    static createFrom(source: any = {}) {
	        return new NewsItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.time = source["time"];
	        this.source = source["source"];
	        this.importance = source["importance"];
	    }
	}
	export class Position {
	    id: number;
	    stockCode: string;
	    stockName: string;
	    buyPrice: number;
	    buyDate: string;
	    quantity: number;
	    costPrice: number;
	    targetPrice: number;
	    stopLossPrice: number;
	    notes: string;
	    status: string;
	    sellPrice: number;
	    sellDate: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Position(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.buyPrice = source["buyPrice"];
	        this.buyDate = source["buyDate"];
	        this.quantity = source["quantity"];
	        this.costPrice = source["costPrice"];
	        this.targetPrice = source["targetPrice"];
	        this.stopLossPrice = source["stopLossPrice"];
	        this.notes = source["notes"];
	        this.status = source["status"];
	        this.sellPrice = source["sellPrice"];
	        this.sellDate = source["sellDate"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class ProAnalysisCache {
	    id: number;
	    stockCode: string;
	    analysisType: string;
	    masterStyle: string;
	    content: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ProAnalysisCache(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.stockCode = source["stockCode"];
	        this.analysisType = source["analysisType"];
	        this.masterStyle = source["masterStyle"];
	        this.content = source["content"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	
	export class ResearchReport {
	    title: string;
	    stockName: string;
	    orgName: string;
	    publishDate: string;
	    researcher: string;
	    rating: string;
	    infoCode: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new ResearchReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.stockName = source["stockName"];
	        this.orgName = source["orgName"];
	        this.publishDate = source["publishDate"];
	        this.researcher = source["researcher"];
	        this.rating = source["rating"];
	        this.infoCode = source["infoCode"];
	        this.url = source["url"];
	    }
	}
	export class Stock {
	    id: number;
	    code: string;
	    name: string;
	    market: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Stock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.market = source["market"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class StockAlert {
	    id: number;
	    stockCode: string;
	    stockName: string;
	    alertType: string;
	    targetValue: number;
	    condition: string;
	    enabled: boolean;
	    triggered: boolean;
	    // Go type: time
	    triggeredAt?: any;
	    triggeredPrice: number;
	    triggeredChange: number;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new StockAlert(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.alertType = source["alertType"];
	        this.targetValue = source["targetValue"];
	        this.condition = source["condition"];
	        this.enabled = source["enabled"];
	        this.triggered = source["triggered"];
	        this.triggeredAt = this.convertValues(source["triggeredAt"], null);
	        this.triggeredPrice = source["triggeredPrice"];
	        this.triggeredChange = source["triggeredChange"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class StockNotice {
	    title: string;
	    date: string;
	    type: string;
	    stockName: string;
	    artCode: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new StockNotice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.date = source["date"];
	        this.type = source["type"];
	        this.stockName = source["stockName"];
	        this.artCode = source["artCode"];
	        this.url = source["url"];
	    }
	}
	export class TradeLevelDetail {
	    buy: number;
	    sell: number;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new TradeLevelDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.buy = source["buy"];
	        this.sell = source["sell"];
	        this.reason = source["reason"];
	    }
	}
	export class TradeLevelResult {
	    shortTerm: TradeLevelDetail;
	    midTerm: TradeLevelDetail;
	    longTerm: TradeLevelDetail;
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new TradeLevelResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shortTerm = this.convertValues(source["shortTerm"], TradeLevelDetail);
	        this.midTerm = this.convertValues(source["midTerm"], TradeLevelDetail);
	        this.longTerm = this.convertValues(source["longTerm"], TradeLevelDetail);
	        this.raw = source["raw"];
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
	export class USStock {
	    id: number;
	    symbol: string;
	    name: string;
	    nameCn: string;
	    exchange: string;
	    sector: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new USStock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.symbol = source["symbol"];
	        this.name = source["name"];
	        this.nameCn = source["nameCn"];
	        this.exchange = source["exchange"];
	        this.sector = source["sector"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class UpdateInfo {
	    hasUpdate: boolean;
	    version: string;
	    currentVersion: string;
	    description: string;
	    downloadUrl: string;
	    releaseUrl: string;
	    releaseDate: string;
	    skipVersion: string;
	    skipped: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasUpdate = source["hasUpdate"];
	        this.version = source["version"];
	        this.currentVersion = source["currentVersion"];
	        this.description = source["description"];
	        this.downloadUrl = source["downloadUrl"];
	        this.releaseUrl = source["releaseUrl"];
	        this.releaseDate = source["releaseDate"];
	        this.skipVersion = source["skipVersion"];
	        this.skipped = source["skipped"];
	    }
	}
	export class VersionInfo {
	    version: string;
	    buildTime: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.buildTime = source["buildTime"];
	    }
	}

}

export namespace plugin {
	
	export class AIChatMessage {
	    role: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new AIChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	    }
	}
	export class AIConfig {
	    provider: string;
	    baseUrl: string;
	    apiKey: string;
	    model: string;
	    maxTokens: number;
	    temperature: number;
	    systemPrompt: string;
	    headers: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new AIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.baseUrl = source["baseUrl"];
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
	        this.maxTokens = source["maxTokens"];
	        this.temperature = source["temperature"];
	        this.systemPrompt = source["systemPrompt"];
	        this.headers = source["headers"];
	    }
	}
	export class DatasourceResult {
	    price: number;
	    change: number;
	    changePercent: number;
	    volume: number;
	    amount: number;
	    high: number;
	    low: number;
	    open: number;
	    preClose: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new DatasourceResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.price = source["price"];
	        this.change = source["change"];
	        this.changePercent = source["changePercent"];
	        this.volume = source["volume"];
	        this.amount = source["amount"];
	        this.high = source["high"];
	        this.low = source["low"];
	        this.open = source["open"];
	        this.preClose = source["preClose"];
	        this.name = source["name"];
	    }
	}
	export class NotificationConfig {
	    url: string;
	    method: string;
	    headers: Record<string, string>;
	    params: Record<string, string>;
	    bodyTemplate: any;
	    contentType: string;
	
	    static createFrom(source: any = {}) {
	        return new NotificationConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.method = source["method"];
	        this.headers = source["headers"];
	        this.params = source["params"];
	        this.bodyTemplate = source["bodyTemplate"];
	        this.contentType = source["contentType"];
	    }
	}
	export class NotificationData {
	    stockCode: string;
	    stockName: string;
	    alertType: string;
	    currentPrice: number;
	    condition: string;
	    targetValue: number;
	    triggerTime: string;
	    change: number;
	    changePercent: number;
	
	    static createFrom(source: any = {}) {
	        return new NotificationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.alertType = source["alertType"];
	        this.currentPrice = source["currentPrice"];
	        this.condition = source["condition"];
	        this.targetValue = source["targetValue"];
	        this.triggerTime = source["triggerTime"];
	        this.change = source["change"];
	        this.changePercent = source["changePercent"];
	    }
	}
	export class NotificationTemplate {
	    id: string;
	    name: string;
	    description: string;
	    config: NotificationConfig;
	
	    static createFrom(source: any = {}) {
	        return new NotificationTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.config = this.convertValues(source["config"], NotificationConfig);
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
	export class Plugin {
	    id: string;
	    name: string;
	    type: string;
	    version: string;
	    author: string;
	    description: string;
	    homepage: string;
	    enabled: boolean;
	    config: number[];
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Plugin(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.description = source["description"];
	        this.homepage = source["homepage"];
	        this.enabled = source["enabled"];
	        this.config = source["config"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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

export namespace prompt {
	
	export class IndicatorResult {
	    signal: string;
	    value: number;
	    text: string;
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new IndicatorResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.signal = source["signal"];
	        this.value = source["value"];
	        this.text = source["text"];
	        this.raw = source["raw"];
	    }
	}
	export class PromptInfo {
	    name: string;
	    type: string;
	    content: string;
	    filePath: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new PromptInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.content = source["content"];
	        this.filePath = source["filePath"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class StockReview {
	    code: string;
	    name: string;
	    action: string;
	    reason: string;
	    targetPrice: number;
	
	    static createFrom(source: any = {}) {
	        return new StockReview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.action = source["action"];
	        this.reason = source["reason"];
	        this.targetPrice = source["targetPrice"];
	    }
	}
	export class ReviewResult {
	    summary: string;
	    performance: string;
	    suggestions: string[];
	    stockReviews: StockReview[];
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new ReviewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary = source["summary"];
	        this.performance = source["performance"];
	        this.suggestions = source["suggestions"];
	        this.stockReviews = this.convertValues(source["stockReviews"], StockReview);
	        this.raw = source["raw"];
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
	export class ScreenerStock {
	    code: string;
	    name: string;
	    reason: string;
	    signal: string;
	
	    static createFrom(source: any = {}) {
	        return new ScreenerStock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.reason = source["reason"];
	        this.signal = source["signal"];
	    }
	}
	export class ScreenerResult {
	    stocks: ScreenerStock[];
	    summary: string;
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new ScreenerResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stocks = this.convertValues(source["stocks"], ScreenerStock);
	        this.summary = source["summary"];
	        this.raw = source["raw"];
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

export namespace struct { ID string "json:\"id\""; Name string "json:\"name\""; Description string "json:\"description\""; Config plugin {
	
	export class  {
	    id: string;
	    name: string;
	    description: string;
	    config: plugin.AIConfig;
	
	    static createFrom(source: any = {}) {
	        return new (source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.config = this.convertValues(source["config"], plugin.AIConfig);
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

export namespace struct { Type prompt {
	
	export class  {
	    type: string;
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new (source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}

}

