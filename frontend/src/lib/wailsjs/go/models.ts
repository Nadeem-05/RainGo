export namespace main {
	
	export class Entry {
	    id: number;
	    pwd: string;
	    hash: string;
	    type: string;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.pwd = source["pwd"];
	        this.hash = source["hash"];
	        this.type = source["type"];
	        this.source = source["source"];
	    }
	}
	export class HashStats {
	    ID: number;
	    SuccessRate: number;
	    FailureRate: number;
	    // Go type: time
	    UpdatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new HashStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.SuccessRate = source["SuccessRate"];
	        this.FailureRate = source["FailureRate"];
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
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

