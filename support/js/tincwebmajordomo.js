export class TincWebMajordomoError extends Error {
    constructor(message, code, details) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}

export class TincWebMajordomo {
    /**
    Operations for joining public network
    **/

    // Create new API handler to TincWebMajordomo.
    // preflightHandler (if defined) can return promise
    constructor(base_url = 'http://127.0.0.1:8686/api/', preflightHandler = null) {
        this.__url = base_url;
        this.__id = 1;
        this.__preflightHandler = preflightHandler;
    }


    /**
    Join public network if code matched. Will generate error if node subnet not matched
    **/
    async join(network, self){
        return (await this.__call('Join', {
            "jsonrpc" : "2.0",
            "method" : "TincWebMajordomo.Join",
            "id" : this.__next_id(),
            "params" : [network, self]
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
            throw new TincWebMajordomoError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}