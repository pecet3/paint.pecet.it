package api

import (
	"log"
	"net/http"

	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wsmanager"
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

// func (api *Api) Run() {
// 	wsM := wsmanager.New()
// 	api.wsM = wsM

// 	http.HandleFunc("/ws", api.handlePaintWS)
// 	log.Println("Listening on port", port)
// 	log.Fatal(http.ListenAndServe(port, nil))
// }

func (api *Api) Run() {
	wsM := wsmanager.New()
	api.wsM = wsM
	w := ward.New()
	w.HandleFunc("/ws", api.handlePaintWS)

	w.Handle("/test", func(greq *ward.Request) {
		greq.Write([]byte("aaaa"))
	})
	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(port, w))
}

func (api *Api) HandleHelloWorld(req *ward.Request) {

}
