export type RoomInfo = {
	name: string; // `json:"name"`
	is_temporary: boolean; // `json:"is_temporary"`
	online_users: number; // `json:"online_users"`
	is_password: boolean; // `json:"is_password"`
	width: number; // `json:"width"`
	height: number; // `json:"height"`
}

export type RoomConfig = {
	name: string; // `json:"name" validate:"required,min=3,max=32"`
	is_temporary?: boolean; // `json:"is_temporary" validate:"omitempty"`
	password?: string; // `json:"password" validate:"omitempty,min=4,max=64"`
	width: number; // `json:"width" validate:"required,gte=100,lte=10000"`
	height: number; // `json:"height" validate:"required,gte=100,lte=10000"`
	is_webrtc?: boolean; // `json:"is_webrtc" validate:"omitempty"`
	is_synth?: boolean; // `json:"is_synth" validate:"omitempty"`
}

export type ServerMessage = {
	message: string; // `json:"message"`
	date: string; // `json:"date"`
}

export type SignalPayload = {
	targetUuid: string; // `json:"targetUuid"`
	senderUuid: string; // `json:"senderUuid"`
	signalType: string; // `json:"signalType"`
	data: any; // `json:"data"`
}

export type UserManagmentPayload = {
	uuid: string; // `json:"uuid"`
}

export type LoginRequest = {
	name: string; // `json:"name" validate:"required,min=2,max=32"`
	password?: string; // `json:"password,omitempty"`
}

export type Event = {
	type: string; // `json:"type"`
	payload: any; // `json:"payload"`
}

export type ChatMessage = {
	name: string; // `json:"name"`
	uuid: string; // `json:"uuid"`
	message: string; // `json:"message"`
	date: string; // `json:"date"`
}

export type RoomUser = {
	uuid: string; // `json:"uuid"`
	name: string; // `json:"name"`
	is_operator: boolean; // `json:"is_operator"`
	is_connected: boolean; // `json:"is_connected"`
	is_draw: boolean; // `json:"is_draw"`
	is_kicked: boolean; // `json:"is_kicked"`
}

export type User = {
	uuid: string; // `json:"uuid"`
	name: string; // `json:"name"`
	rank: number; // `json:"rank"`
}

export type ByteEvent = {
	type: string; // `json:"type"`
	payload: any[]; // `json:"payload"`
}

