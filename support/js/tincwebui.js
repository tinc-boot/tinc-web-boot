export class TincWebUIError extends Error {
    constructor(message, code, details) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}

export class TincWebUI {
    /**
    Operations with tinc-web-boot related to UI
    **/

    // Create new API handler to TincWebUI.
    // preflightHandler (if defined) can return promise
    constructor(base_url = 'http://127.0.0.1:8686/api/', preflightHandler = null) {
        this.__url = base_url;
        this.__id = 1;
        this.__preflightHandler = preflightHandler;
    }


    /**
    Issue and sign token
    **/
    async issueAccessToken(validDays){
        return (await this.__call('IssueAccessToken', {
            "jsonrpc" : "2.0",
            "method" : "TincWebUI.IssueAccessToken",
            "id" : this.__next_id(),
            "params" : [validDays]
        }));
    }

    /**
    Make desktop notification if system supports it
    **/
    async notify(title, message){
        return (await this.__call('Notify', {
            "jsonrpc" : "2.0",
            "method" : "TincWebUI.Notify",
            "id" : this.__next_id(),
            "params" : [title, message]
        }));
    }

    /**
    Endpoints list to access web UI
    **/
    async endpoints(){
        return (await this.__call('Endpoints', {
            "jsonrpc" : "2.0",
            "method" : "TincWebUI.Endpoints",
            "id" : this.__next_id(),
            "params" : []
        }));
    }



    __next_id() {
        this.__id += 1;
        return this.__id
    }

    async __call(method, req) {
        const fetchParams = {
            method: "POST",
            headers: {
                'Content-Type' : 'application/json',
            },
            body: JSON.stringify(req)
        };
        if (this.__preflightHandler) {
            await Promise.resolve(this.__preflightHandler(method, fetchParams));
        }
        const res = await fetch(this.__url, fetchParams);
        if (!res.ok) {
            throw new Error(res.status + ' ' + res.statusText);
        }

        const data = await res.json();

        if ('error' in data) {
            throw new TincWebUIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}