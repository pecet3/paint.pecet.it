export interface User {
    uuid: string;
    name: string;
    rank: number;
}
export type Event = {
    type: string;
    payload: any;
}

export interface NoteEvent {
    note: string;
    type: 'attack' | 'release';
    synthType: SynthType;
    userId: string;
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

export type RoomUser = {
    uuid: string;
    name: string;
    is_operator: boolean;
    is_connected: boolean;
    is_draw: boolean;
    is_kicked: boolean;
}

export interface WebRTCSignalPayload {
    targetUuid: string;
    senderUuid: string;
    signalType: "offer" | "answer" | "ice";
    data: any;
}

export type RoomConfig = {
    name: string;
    password: string;
    is_temporary: boolean;
    width: number;
    height: number;
};

export type RoomInfo = {
    name: string;
    online_users: number;
    is_temporary: boolean;
    is_passowrd: boolean;
    width: number;
    height: number;
};