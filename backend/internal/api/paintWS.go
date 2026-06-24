package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"paint.pecet.it/pkg/guardian"
	"paint.pecet.it/pkg/paint"
	"paint.pecet.it/pkg/wsmanager"
)

const bufSize = 1024 * 64 * 4

var upgrader = websocket.Upgrader{
	ReadBufferSize:  bufSize,
	WriteBufferSize: bufSize,
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
	greq := guardian.GetGuardianRequest(r)

	room, ok := api.wsM.GetRoom(roomName)
	if !ok {
		room = wsmanager.NewRoom(roomName)
		p := paint.New(room, 800, 600)
		p.Run()
		api.wsM.SetRoom(room, roomName)
		room.Run()
	}

	room.HandleNewClient(conn, greq)
}
