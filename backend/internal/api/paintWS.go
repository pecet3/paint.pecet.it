package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"paint.pecet.it/pkg/paint"
	"paint.pecet.it/pkg/wsmanager"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (api *Api) handlePaintWS(w http.ResponseWriter, r *http.Request) {
	roomName := r.URL.Query().Get("room")
	if roomName == "" {
		roomName = "general"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	room, ok := api.wsM.GetRoom(roomName)
	if !ok {
		room = wsmanager.NewRoom()
		paint.ImplementPaint(room)
		api.wsM.SetRoom(room, roomName)
		go room.Run()
	}

	room.HandleNewConn(conn)
}
