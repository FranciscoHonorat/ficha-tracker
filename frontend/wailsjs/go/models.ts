export namespace main {
	
	export class FichaView {
	    id: string;
	    fullName: string;
	    requestType: string;
	    acs: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new FichaView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fullName = source["fullName"];
	        this.requestType = source["requestType"];
	        this.acs = source["acs"];
	        this.createdAt = source["createdAt"];
	    }
	}

}

