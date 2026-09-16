export namespace main {
	
	export class ACSStatView {
	    acsId: string;
	    acsName: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new ACSStatView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.acsId = source["acsId"];
	        this.acsName = source["acsName"];
	        this.count = source["count"];
	    }
	}
	export class ACSView {
	    id: string;
	    name: string;
	    phone: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ACSView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.phone = source["phone"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class FichaView {
	    id: string;
	    fullName: string;
	    requestType: string;
	    acsId: string;
	    acsName: string;
	    phone: string;
	    notified: boolean;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new FichaView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fullName = source["fullName"];
	        this.requestType = source["requestType"];
	        this.acsId = source["acsId"];
	        this.acsName = source["acsName"];
	        this.phone = source["phone"];
	        this.notified = source["notified"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class RequestTypeStatView {
	    requestType: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new RequestTypeStatView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestType = source["requestType"];
	        this.count = source["count"];
	    }
	}

}

