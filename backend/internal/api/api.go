package api

import (
	"log"
	"net/http"

	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
)

const (
	port = ":8080"
)

type Api struct {
	wardsocket *wardsocket.WardSocket
}

func New() *Api {
	return &Api{}
}

// func (api *Api) Run() {
// 	wsM := wardsocket.New()
// 	api.wsM = wsM

// 	http.HandleFunc("/ws", api.handlePaintWS)
// 	log.Println("Listening on port", port)
// 	log.Fatal(http.ListenAndServe(port, nil))
// }

func (api *Api) Run() {
	w := ward.New()
	wardsocket := wardsocket.New(w)
	api.wardsocket = wardsocket

	w.Handle("/ws", api.handlePaintWS)

	w.Handle("/test", api.handleHelloWorld)

	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(port, w))
}

func (api *Api) handleHelloWorld(wreq *ward.Request) {
	wreq.Write([]byte("hello world"))
}
