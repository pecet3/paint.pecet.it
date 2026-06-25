package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"paint.pecet.it/pkg/paintroom"
	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
)

const bufSize = 1024 * 64 * 4

const generalRoomIdent = "1"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  bufSize,
	WriteBufferSize: bufSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (api *Api) handlePaintWS(wreq *ward.Request) {
	conn, err := upgrader.Upgrade(wreq.ResponseWriter, wreq.Http, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	room, ok := api.wardsocket.GetRoom(generalRoomIdent)
	if !ok {
		r, ctx := wardsocket.NewRoom(generalRoomIdent).WithContext()
		room = r
		paintroom.New(room).Run(ctx)
		api.wardsocket.AddRoom(room)
		room.Run(ctx)
	}

	room.HandleNewClient(conn, wreq)
}
