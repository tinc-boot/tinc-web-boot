
export interface NetworkID {
    name: string
}

export interface PeerID {
    network: string
    node: string
    subnet: string | null
}


export type EventName = 'Started' | 'Stopped' | 'PeerDiscovered' | 'PeerJoined' | 'PeerLeft';
export type EventPayload = NetworkID | PeerID;
export type EventHandler = ((payload: EventPayload, event: EventName) => (void)) | ((payload: EventPayload) => (void))

export class Events {
    private stopped = false;
    private readonly listeners = new Map<EventName, EventHandler>();

    constructor(private readonly url: string, private readonly reconnectInterval: number = 1000) {
        this.start();
    }

    on(event: EventName, handler: EventHandler) {
        this.listeners.set(event, handler)
    }

    off(event: EventName, handler: EventHandler) {
        this.listeners.delete(event)
    }


    onStarted(handler: ((payload: NetworkID) => (void)) | ((payload: NetworkID, event: EventName) => (void))) {
        this.listeners.set('Started', handler as EventHandler);
    }

    offStarted(handler: ((payload: NetworkID) => (void)) | ((payload: NetworkID, event: EventName) => (void))) {
        this.listeners.delete('Started');
    }


    onStopped(handler: ((payload: NetworkID) => (void)) | ((payload: NetworkID, event: EventName) => (void))) {
        this.listeners.set('Stopped', handler as EventHandler);
    }

    offStopped(handler: ((payload: NetworkID) => (void)) | ((payload: NetworkID, event: EventName) => (void))) {
        this.listeners.delete('Stopped');
    }


    onPeerDiscovered(handler: ((payload: PeerID) => (void)) | ((payload: PeerID, event: EventName) => (void))) {
        this.listeners.set('PeerDiscovered', handler as EventHandler);
    }

    offPeerDiscovered(handler: ((payload: PeerID) => (void)) | ((payload: PeerID, event: EventName) => (void))) {
        this.listeners.delete('PeerDiscovered');
    }


    onPeerJoined(handler: ((payload: PeerID) => (void)) | ((payload: PeerID, event: EventName) => (void))) {
        this.listeners.set('PeerJoined', handler as EventHandler);
    }

    offPeerJoined(handler: ((payload: PeerID) => (void)) | ((payload: PeerID, event: EventName) => (void))) {
        this.listeners.delete('PeerJoined');
    }


    onPeerLeft(handler: ((payload: PeerID) => (void)) | ((payload: PeerID, event: EventName) => (void))) {
        this.listeners.set('PeerLeft', handler as EventHandler);
    }

    offPeerLeft(handler: ((payload: PeerID) => (void)) | ((payload: PeerID, event: EventName) => (void))) {
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
                setTimeout(() => this.start(), this.reconnectInterval);
            }
        }
        socket.onerror = (e) => {
            console.error(e);
            if (!restarted) {
                restarted = true;
                socket.close();
                setTimeout(() => this.start(), this.reconnectInterval);
            }
        }
        socket.onmessage = ({data}) => {
            const {event, payload} = JSON.parse(data) as { event: string, payload: any };
            const handler = this.listeners.get(event as EventName);
            if (handler) {
                try{
                    handler(payload, event as EventName);
                } catch(e) {
                    console.error(`failed to process handler for event ${event}:`, e);
                }
            }
        }
    }

}