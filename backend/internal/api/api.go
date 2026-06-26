package api

import (
	"log"
	"net/http"

	"paint.pecet.it/internal/simpleauth"
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
	w := ward.New("/api")
	wardsocket := wardsocket.New(w, &wardsocket.Upgrader{
		ReadBufferSize:  bufSize,
		WriteBufferSize: bufSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})

	api.wardsocket = wardsocket
	auth := simpleauth.New()

	public := w.NewGroup("")
	public.Post("/login", auth.LoginHandler)
	public.Get("/hello", api.handleHelloWorld, api.someMiddleware(100))

	protected := w.NewGroup("")
	protected.Use(auth.AuthMiddleware)

	protected.Get("/ws", api.handlePaintWS)
	protected.Get("/ping", auth.PingHandler)

	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(port, w))
}

func (api *Api) handleHelloWorld(wreq *ward.Request) {
	wreq.Write([]byte("hello world"))
}

func (api *Api) someMiddleware(perms int) ward.Middleware {
	return func(next ward.Handler) ward.Handler {
		return func(wreq *ward.Request) {
			log.Println("[Middleware] Sprawdzam żądanie przed uruchomieniem handlera...", perms)

			next(wreq)
		}
	}
}
