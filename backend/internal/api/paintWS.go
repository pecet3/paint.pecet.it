package api

import (
	"paint.pecet.it/pkg/paintroom"
	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
)

const generalRoomIdent = "1"

func (api *Api) handlePaintWS(wreq *ward.Request) {
	room, ok := api.wardsocket.GetRoom(generalRoomIdent)
	if !ok {
		r, ctx := wardsocket.NewRoom(generalRoomIdent).WithContext()
		room = r
		paintroom.New(room).Run(ctx)
		api.wardsocket.AddRoom(room)
		room.Run(ctx)
	}

	err := api.wardsocket.AssignRequestToRoom(wreq, room)
	if err != nil {
		wreq.Log(err)
		return
	}
}
