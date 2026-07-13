package api

import (
	"encoding/json"
	"log"

	"paint.pecet.it/internal/pkg/paint"
	"paint.pecet.it/pkg/ward"
)

func (api *Api) handleJoinRoom(wreq *ward.Request) {
	roomName := wreq.Http.PathValue("name")
	if err := api.paint.AssignRequestToRoom(wreq, roomName); err != nil {
		wreq.WriteErr(404, "Room not found")
		return
	}
}
func (api *Api) handleGetRoom(wreq *ward.Request) {
	roomName := wreq.Http.PathValue("name")
	r := api.paint.GetRoom(roomName)
	if r != nil {
		wreq.WriteJson(r.Info())
	} else {
		wreq.WriteErr(404, "Room not found")
		return
	}
}

func (api *Api) handleCreateRoom(wreq *ward.Request) {
	var cfg paint.RoomConfig
	err := json.NewDecoder(wreq.Http.Body).Decode(&cfg)
	if err != nil {
		log.Println(err, cfg)
		wreq.WriteErr(400, "Invalid JSON")
		return
	}
	if r := api.paint.GetRoom(cfg.Name); r != nil {
		wreq.WriteErr(400, "Room already exists")
		return
	}
	cfg.IsTemporary = true
	api.paint.CreateRoom(&cfg)
}

func (api *Api) handleRoomsList(wreq *ward.Request) {
	rooms := api.paint.ListRooms()
	wreq.WriteJson(rooms)
}
