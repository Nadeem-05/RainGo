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

}

