export type ByteEvent = {
	type: string;
	payload: any[];
}

export type ServerMessage = {
	message: string;
	date: string;
}

export type User = {
	uuid: string;
	name: string;
	rank: number;
}

export type LoginRequest = {
	name: string;
	password: string;
}

export type RoomInfo = {
	name: string;
	is_temporary: boolean;
	online_users: number;
	is_password: boolean;
	width: number;
	height: number;
}

export type RoomConfig = {
	name: string;
	is_temporary: boolean;
	password: string;
	width: number;
	height: number;
}

export type ChatMessage = {
	name: string;
	uuid: string;
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

export type SignalPayload = {
	targetUuid: string;
	senderUuid: string;
	signalType: string;
	data: any;
}

export type UserManagmentPayload = {
	uuid: string;
}

export type Event = {
	type: string;
	payload: any;
}

