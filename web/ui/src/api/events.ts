namespace EventTypes {
    
    export interface NetworkID {
        name: string
    }
    
}

export interface Events {
    
    onStarted(handler: (payload:EventTypes.NetworkID)=>(void));
    
    onStopped(handler: (payload:EventTypes.NetworkID)=>(void));
    
}