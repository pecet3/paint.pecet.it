export type PaintUserManagmentPayload = {
	uuid: string; // `json:"uuid"`
}

export type SimpleauthUser = {
	uuid: string; // `json:"uuid"`
	name: string; // `json:"name"`
	rank: number; // `json:"rank"`
}

export type SimpleauthLoginRequest = {
	name: string; // `json:"name" validate:"required,min=2,max=32"`
	password: string; // `json:"password,omitempty"`
}

export type PaintRoomConfig = {
	name: string; // `json:"name" validate:"required,min=3,max=32"`
	is_temporary: boolean; // `json:"is_temporary" validate:"omitempty"`
	password: string; // `json:"password" validate:"omitempty,min=4,max=64"`
	width: number; // `json:"width" validate:"required,gte=100,lte=10000"`
	height: number; // `json:"height" validate:"required,gte=100,lte=10000"`
}

export type PaintMessage = {
	message_test: test[]; // `json:"message_test"`
}

export type WardsocketEvent = {
	type: string; // `json:"type"`
	payload: any; // `json:"payload"`
}

export type WardsocketByteEvent = {
	type: string; // `json:"type"`
	payload: any[]; // `json:"payload"`
}

export type PaintPaintRoomInfo = {
	name: string; // `json:"name"`
	is_temporary: boolean; // `json:"is_temporary"`
	online_users: number; // `json:"online_users"`
	is_password: boolean; // `json:"is_password"`
	width: number; // `json:"width"`
	height: number; // `json:"height"`
}

export type PaintChatMessage = {
	name: string; // `json:"name"`
	uuid: string; // `json:"uuid"`
	message: string; // `json:"message"`
	date: string; // `json:"date"`
}

// name: test
export type test = {
	message: string; // `json:"message"`
	date: string; // `json:"date"`
}

export type PaintRoomUser = {
	uuid: string; // `json:"uuid"`
	name: string; // `json:"name"`
	is_operator: boolean; // `json:"is_operator"`
	is_connected: boolean; // `json:"is_connected"`
	is_drawing: boolean; // `json:"is_drawing"`
	is_kicked: boolean; // `json:"is_kicked"`
}

export type PaintSignalPayload = {
	targetUuid: string; // `json:"targetUuid"`
	senderUuid: string; // `json:"senderUuid"`
	signalType: string; // `json:"signalType"`
	data: any; // `json:"data"`
}

