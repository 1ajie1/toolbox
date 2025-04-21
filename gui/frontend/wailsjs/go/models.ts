export namespace network {
	
	export class DNSResult {
	    success: boolean;
	    domain: string;
	    ipList: string[];
	    error: string;
	    serverUsed: string;
	
	    static createFrom(source: any = {}) {
	        return new DNSResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.domain = source["domain"];
	        this.ipList = source["ipList"];
	        this.error = source["error"];
	        this.serverUsed = source["serverUsed"];
	    }
	}
	export class PingResult {
	    success: boolean;
	    avgLatency: string;
	    packetLoss: string;
	    error: string;
	    outputLines: string[];
	
	    static createFrom(source: any = {}) {
	        return new PingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.avgLatency = source["avgLatency"];
	        this.packetLoss = source["packetLoss"];
	        this.error = source["error"];
	        this.outputLines = source["outputLines"];
	    }
	}
	export class SpeedTestConfig {
	    port: number;
	    host: string;
	    dataSize: number;
	
	    static createFrom(source: any = {}) {
	        return new SpeedTestConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.host = source["host"];
	        this.dataSize = source["dataSize"];
	    }
	}

}

