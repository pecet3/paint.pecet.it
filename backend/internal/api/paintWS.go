package api

import (
	"context"

	"paint.pecet.it/internal/paintroom"
	"paint.pecet.it/internal/usermanager"
	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
	"paint.pecet.it/pkg/ward/wardsocket/webrtc"
)

const generalRoomIdent = "1"

func (api *Api) handlePaintWS(wreq *ward.Request) {
	room, ok := api.wardsocket.GetRoom(generalRoomIdent)
	if !ok {
		ctx := context.Background()
		r, ctx := wardsocket.NewRoom(generalRoomIdent).WithCancelContext(ctx)

		usermanager.New(r).RegisterHandlers()
		webrtc.New(r).RegisterHandlers()

		pr := paintroom.New(r)
		pr.RegisterHandlers()
		pr.Run(ctx)

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
