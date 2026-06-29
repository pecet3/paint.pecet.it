package api

import (
	"context"

	"paint.pecet.it/pkg/paintroom"
	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
	"paint.pecet.it/pkg/ward/wardsocket/usermanager"
)

const generalRoomIdent = "1"

func (api *Api) handlePaintWS(wreq *ward.Request) {
	room, ok := api.wardsocket.GetRoom(generalRoomIdent)
	if !ok {
		ctx := context.Background()
		r, ctx := wardsocket.NewRoom(generalRoomIdent).WithCancelContext(ctx)
		um := usermanager.New(r)
		ctx = um.WithContext(ctx)
		paintroom.New(r).Run(ctx)
		api.wardsocket.AddRoom(r)
		r.Run(ctx)

		room = r
	}

	err := api.wardsocket.AssignRequestToRoom(wreq, room)
	if err != nil {
		wreq.Log(err)
		return
	}
}
