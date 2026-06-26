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

const bufSize = 1024 * 64 * 4

func (api *Api) Run() {
	w := ward.New()
	wardsocket := wardsocket.New(w, &wardsocket.Upgrader{
		ReadBufferSize:  bufSize,
		WriteBufferSize: bufSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})

	api.wardsocket = wardsocket

	w.Get("/ws", api.handlePaintWS)

	w.Get("/test", api.handleHelloWorld)

	w.Post("/test", api.handleHelloWorld)
	w.Delete("/test", api.handleHelloWorld)

	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(port, w))
}

func (api *Api) handleHelloWorld(wreq *ward.Request) {
	wreq.Write([]byte("hello world"))
}
