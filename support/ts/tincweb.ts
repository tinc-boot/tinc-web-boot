export class TincWebError extends Error {
    public readonly code: number;
    public readonly details: any;

    constructor(message: string, code: number, details: any) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}


export interface Network {
    name: string
    running: boolean
    config: Config | null
}

export interface Config {
    name: string
    port: number
    interface: string
    autostart: boolean
    mode: string
    ip: string
    mask: number
    deviceType: string | null
    device: string | null
    connectTo: Array<string> | null
}

export interface PeerInfo {
    name: string
    online: boolean
    status: Peer | null
    config: Node | null
}

export interface Peer {
    address: string
    fetched: boolean
    config: Node | null
}

export interface Node {
    name: string
    subnet: string
    port: number
    address: Array<Address> | null
    publicKey: string
    version: number
}

export interface Address {
    host: string
    port: number | null
}

export interface Sharing {
    name: string
    subnet: string
    node: Array<Node> | null
}

export interface Upgrade {
    port: number | null
    address: Array<Address> | null
    device: string | null
}



export type Duration = string; // suffixes: ns, us, ms, s, m, h (!!!!)


// support stuff


interface rpcExecutor {
    call(id: number, payload: string): Promise<object>;
}

class wsExecutor {
    private socket?: WebSocket;
    private connecting = false;
    private readonly pendingConnection: Array<() => (void)> = [];
    private readonly correlation = new Map<number, [(data: object) => void, (err: object) => void]>();

    constructor(private readonly url: string) {
    }

    async call(id: number, payload: string): Promise<object> {
        const conn = await this.connectIfNeeded();
        if (this.correlation.has(id)) {
            throw new Error(`already exists pending request with id ${id}`);
        }
        let future = new Promise<object>((resolve, reject) => {
            this.correlation.set(id, [resolve, reject]);
        });
        conn.send(payload);
        return (await future);
    }

    private async connectIfNeeded(): Promise<WebSocket> {
        while (this.connecting) {
            await new Promise((resolve => {
                this.pendingConnection.push(resolve);
            }))
        }
        if (this.socket) {
            return this.socket;
        }
        this.connecting = true;
        let socket;
        try {
            socket = await this.connect();
        } finally {
            this.connecting = false;
        }
        socket.onerror = () => {
            this.onConnectionFailed();
        }
        socket.onclose = () => {
            this.onConnectionFailed();
        }
        socket.onmessage = ({data}) => {
            let res;
            try {
                res = JSON.parse(data);
            } catch (e) {
                console.error("failed parse request:", e);
            }
            const task = this.correlation.get(res.id);
            if (task) {
                this.correlation.delete(res.id);
                task[0](res);
            }
        }
        this.socket = socket;

        let cp = this.pendingConnection;
        this.pendingConnection.slice(0, 0);
        cp.forEach((f) => f());
        return this.socket;
    }

    private connect(): Promise<WebSocket> {
        return new Promise<WebSocket>(((resolve, reject) => {
            let socket = new WebSocket(this.url);
            let resolved = false;
            socket.onopen = () => {
                resolved = true;
                resolve(socket);
            }

            socket.onerror = (e) => {
                if (!resolved) {
                    reject(e);
                    resolved = true;
                }
            }

            socket.onclose = (e) => {
                if (!resolved) {
                    reject(e);
                    resolved = true;
                }
            }
        }));
    }

    private onConnectionFailed() {
        let sock = this.socket;
        this.socket = undefined;
        if (sock) {
            sock.close();
        }
        const cp = Array.from(this.correlation.values());
        this.correlation.clear();
        const err = new Error('connection closed');
        cp.forEach((([_, reject]) => {
            reject(err);
        }))
    }
}

class postExecutor {
    constructor(private readonly url: string) {
    }

    async call(id: number, payload: string): Promise<object> {
        const fetchParams = {
            method: "POST",
            headers: {
                'Content-Type': 'application/json',
            },
            body: payload
        };
        const res = await fetch(this.url, fetchParams);
        if (!res.ok) {
            throw new Error(res.status + ' ' + res.statusText);
        }
        return await res.json();
    }
}

/**
Public Tinc-Web API (json-rpc 2.0)
**/
export class TincWeb {

    private __id: number;
    private __executor:rpcExecutor;


    // Create new API handler to TincWeb.
    constructor(base_url : string = 'ws://127.0.0.1:8686/api/') {
        const proto = (new URL(base_url)).protocol;
        switch (proto) {
            case "ws:":
            case "wss:":{
                this.__executor=new wsExecutor(base_url);
                break
            }
            case "http:":
            case "https:":
            default:{
                this.__executor = new postExecutor(base_url);
                break
            }
        }
        this.__id = 1;
    }


    /**
    List of available networks (briefly, without config)
    **/
    async networks(): Promise<Array<Network>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Networks",
            "id" : this.__next_id(),
            "params" : []
        })) as Array<Network>;
    }

    /**
    Detailed network info
    **/
    async network(name: string): Promise<Network> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Network",
            "id" : this.__next_id(),
            "params" : [name]
        })) as Network;
    }

    /**
    Create new network if not exists
    **/
    async create(name: string, subnet: string): Promise<Network> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Create",
            "id" : this.__next_id(),
            "params" : [name, subnet]
        })) as Network;
    }

    /**
    Remove network (returns true if network existed)
    **/
    async remove(network: string): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Remove",
            "id" : this.__next_id(),
            "params" : [network]
        })) as boolean;
    }

    /**
    Start or re-start network
    **/
    async start(network: string): Promise<Network> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Start",
            "id" : this.__next_id(),
            "params" : [network]
        })) as Network;
    }

    /**
    Stop network
    **/
    async stop(network: string): Promise<Network> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Stop",
            "id" : this.__next_id(),
            "params" : [network]
        })) as Network;
    }

    /**
    Peers brief list in network  (briefly, without config)
    **/
    async peers(network: string): Promise<Array<PeerInfo>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Peers",
            "id" : this.__next_id(),
            "params" : [network]
        })) as Array<PeerInfo>;
    }

    /**
    Peer detailed info by in the network
    **/
    async peer(network: string, name: string): Promise<PeerInfo> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Peer",
            "id" : this.__next_id(),
            "params" : [network, name]
        })) as PeerInfo;
    }

    /**
    Import another tinc-web network configuration file.
It means let nodes defined in config join to the network.
Return created (or used) network with full configuration
    **/
    async import(sharing: Sharing): Promise<Network> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Import",
            "id" : this.__next_id(),
            "params" : [sharing]
        })) as Network;
    }

    /**
    Share network and generate configuration file.
    **/
    async share(network: string): Promise<Sharing> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Share",
            "id" : this.__next_id(),
            "params" : [network]
        })) as Sharing;
    }

    /**
    Node definition in network (aka - self node)
    **/
    async node(network: string): Promise<Node> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Node",
            "id" : this.__next_id(),
            "params" : [network]
        })) as Node;
    }

    /**
    Upgrade node parameters.
In some cases requires restart
    **/
    async upgrade(network: string, update: Upgrade): Promise<Node> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Upgrade",
            "id" : this.__next_id(),
            "params" : [network, update]
        })) as Node;
    }

    /**
    Generate Majordomo request for easy-sharing
    **/
    async majordomo(network: string, lifetime: Duration): Promise<string> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "TincWeb.Majordomo",
            "id" : this.__next_id(),
            "params" : [network, lifetime]
        })) as string;
    }


    private __next_id() {
        this.__id += 1;
        return this.__id
    }

    private async __call(req: { id: number, jsonrpc: string, method: string, params: object | Array<any> }): Promise<any> {
        const data = await this.__executor.call(req.id, JSON.stringify(req)) as {
            error?: {
                message: string,
                code: number,
                data?: any
            },
            result?:any
        }

        if (data.error) {
            throw new TincWebError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}