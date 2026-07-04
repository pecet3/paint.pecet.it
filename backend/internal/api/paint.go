package api

import (
	"encoding/json"
	"log"

	"paint.pecet.it/internal/paint"
	"paint.pecet.it/pkg/ward"
)

func (api *Api) handleJoinRoom(wreq *ward.Request) {
	roomName := wreq.Http.PathValue("id")
	log.Println(roomName)
	if err := api.paint.AssignRequestToRoom(wreq, roomName); err != nil {
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
	api.paint.CreateRoom(&cfg)

}
func (api *Api) handleRoomsList(wreq *ward.Request) {
	rooms := api.paint.ListRooms()
	json.NewEncoder(wreq.ResponseWriter).Encode(rooms)
}
