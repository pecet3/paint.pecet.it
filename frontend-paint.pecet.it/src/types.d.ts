
export type Event = {
    type: string;
    payload: any;
}
export interface Pixel {
    x: number;
    y: number;
    color: string;
}

export interface ChatMessage {
    name: string;
    uuid: string;
    message: string;
    date: string;
}

export interface ServerMessage {
    message: string;
    date: string;
}

export interface RoomUser {
    uuid: string;
    name: string;
    is_operator: boolean;
    is_connected: boolean;
    ban_duration_seconds: number;
}

