export class TincWebError extends Error {
    constructor(message, code, details) {
        super(message);
        this.code = code;
        this.details = details;
    }
}

export class TincWeb {
    /**
    Public Tinc-Web API (json-rpc 2.0)
    **/

    // Create new API handler to TincWeb.
    // preflightHandler (if defined) can return promise
    constructor(base_url = 'http://127.0.0.1:8686/api', preflightHandler = null) {
        this.__url = base_url;
        this.__id = 1;
        this.__preflightHandler = preflightHandler;
    }


    /**
    List of available networks (briefly, without config)
    **/
    async networks(){
        return (await this.__call('Networks', {
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Networks",
            "id" : this.__next_id(),
            "params" : []
        }));
    }

    /**
    Detailed network info
    **/
    async network(name){
        return (await this.__call('Network', {
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Network",
            "id" : this.__next_id(),
            "params" : [name]
        }));
    }

    /**
    Create new network if not exists
    **/
    async create(name){
        return (await this.__call('Create', {
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Create",
            "id" : this.__next_id(),
            "params" : [name]
        }));
    }

    /**
    Remove network (returns true if network existed)
    **/
    async remove(network){
        return (await this.__call('Remove', {
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Remove",
            "id" : this.__next_id(),
            "params" : [network]
        }));
    }

    /**
    Start or re-start network
    **/
    async start(network){
        return (await this.__call('Start', {
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Start",
            "id" : this.__next_id(),
            "params" : [network]
        }));
    }

    /**
    Stop network
    **/
    async stop(network){
        return (await this.__call('Stop', {
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Stop",
            "id" : this.__next_id(),
            "params" : [network]
        }));
    }

    /**
    Peers brief list in network  (briefly, without config)
    **/
    async peers(network){
        return (await this.__call('Peers', {
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Peers",
            "id" : this.__next_id(),
            "params" : [network]
        }));
    }

    /**
    Peer detailed info by in the network
    **/
    async peer(network, name){
        return (await this.__call('Peer', {
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Peer",
            "id" : this.__next_id(),
            "params" : [network, name]
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
            throw new TincWebError(data.message, data.code, data.data);
        }

        return data;
    }
}