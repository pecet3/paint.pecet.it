package paint

// name: RoomInfo
type RoomInfo struct {
	Name        string `json:"name"`
	IsTemporary bool   `json:"is_temporary"`
	OnlineUsers int    `json:"online_users"`
	IsPassword  bool   `json:"is_password"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

// name: RoomConfig
type RoomConfig struct {
	Name        string `json:"name" validate:"required,min=3,max=32"`
	IsTemporary bool   `json:"is_temporary" validate:"omitempty"`
	Password    string `json:"password" validate:"omitempty,min=4,max=64"`
	Width       int    `json:"width" validate:"required,gte=100,lte=10000"`
	Height      int    `json:"height" validate:"required,gte=100,lte=10000"`
	IsWebRTC    bool   `json:"is_webrtc" validate:"omitempty"`
	IsSynth     bool   `json:"is_synth" validate:"omitempty"`
}
