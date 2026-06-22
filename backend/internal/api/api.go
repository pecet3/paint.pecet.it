package api

import (
	"log"
	"net/http"

	"paint.pecet.it/pkg/wsmanager"
)

const (
	port = ":8080"
)

type Api struct {
	wsM *wsmanager.WsManager
}

func New() *Api {
	return &Api{}
}
func (api *Api) Run() {
	wsM := wsmanager.New()
	api.wsM = wsM

	http.HandleFunc("/ws", api.handlePaintWS)
	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
