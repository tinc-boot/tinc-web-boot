namespace EventTypes {
    
    export interface NetworkID {
        name: string
    }
    
    export interface PeerID {
        network: string
        node: string
        subnet: string | null
    }
    
}

export class Events {
    private stopped = true;
    private readonly listeners = new Map<string, (payload: any) => (void)>();

    constructor(private readonly url: string, private readonly reconnectInterval: number = 1000) {
        this.start();
    }


    onStarted(handler: (payload: EventTypes.NetworkID) => (void)) {
        this.listeners['Started'] = handler;
    }

    offStarted(handler: (payload: EventTypes.NetworkID) => (void)) {
        this.listeners.delete('Started');
    }


    onStopped(handler: (payload: EventTypes.NetworkID) => (void)) {
        this.listeners['Stopped'] = handler;
    }

    offStopped(handler: (payload: EventTypes.NetworkID) => (void)) {
        this.listeners.delete('Stopped');
    }


    onPeerDiscovered(handler: (payload: EventTypes.PeerID) => (void)) {
        this.listeners['PeerDiscovered'] = handler;
    }

    offPeerDiscovered(handler: (payload: EventTypes.PeerID) => (void)) {
        this.listeners.delete('PeerDiscovered');
    }


    onPeerJoined(handler: (payload: EventTypes.PeerID) => (void)) {
        this.listeners['PeerJoined'] = handler;
    }

    offPeerJoined(handler: (payload: EventTypes.PeerID) => (void)) {
        this.listeners.delete('PeerJoined');
    }


    onPeerLeft(handler: (payload: EventTypes.PeerID) => (void)) {
        this.listeners['PeerLeft'] = handler;
    }

    offPeerLeft(handler: (payload: EventTypes.PeerID) => (void)) {
        this.listeners.delete('PeerLeft');
    }



    stop() {
        this.stopped = true;
    }

    private start() {
        if (this.stopped) return;
        let restarted = false;
        const socket = new WebSocket(this.url);
        socket.onclose = () => {
            if (!restarted) {
                restarted = true;
                setInterval(() => this.start(), this.reconnectInterval);
            }
        }
        socket.onerror = (e) => {
            console.error(e);
            if (!restarted) {
                restarted = true;
                socket.close();
                setInterval(() => this.start(), this.reconnectInterval);
            }
        }
        socket.onmessage = ({data}) => {
            const {event, payload} = JSON.parse(data) as { event: string, payload: any };
            const handler = this.listeners[event];
            if (handler) {
                try{
                    handler(payload);
                } catch(e) {
                    console.error(`failed to process handler for event ${event}:`, e);
                }
            }
        }
    }

}