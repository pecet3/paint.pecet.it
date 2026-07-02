package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	w := ward.New()
	api.wardsocket = wardsocket.New(w, &wardsocket.Upgrader{
		ReadBufferSize:  bufSize,
		WriteBufferSize: bufSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})
	auth := simpleauth.New()

	public := w.NewGroup("/api")
	public.Post("/login", auth.LoginHandler)
	public.Get("/hello", api.handleHelloWorld)

	protected := w.NewGroup("/api").With(auth.AuthMiddleware)
	protected.Get("/ws", api.handlePaintWS)
	protected.Get("/ping", auth.PingHandler)

	w.Get("/", func(wreq *ward.Request) {
		distPath := "./dist"
		path := filepath.Join(distPath, wreq.Http.URL.Path)
		info, err := os.Stat(path)

		if os.IsNotExist(err) || info.IsDir() {
			http.ServeFile(wreq.ResponseWriter, wreq.Http, filepath.Join(distPath, "index.html"))
			return
		}

		http.ServeFile(wreq.ResponseWriter, wreq.Http, path)
	})
	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(port, w))
}

func (api *Api) handleHelloWorld(wreq *ward.Request) {
	wreq.Write([]byte("hello world"))
}
