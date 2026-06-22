// Typ akcji rysowania
export type DrawAction = 'start' | 'draw' | 'end';

// Reprezentacja pojedynczego punktu
export interface Point {
    x: number;
    y: number;
}

// Główna struktura (Event) przesyłana przez WebSocket
export interface DrawEvent {
    action: DrawAction;
    point: Point;
    color: string;
    size: number;
    // Opcjonalnie: sessionId lub userId, aby ignorować własne echa z serwera
    userId?: string;
}

export type Event = {
    type: string;
    payload: any;
}