package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"paint.pecet.it/internal/repo/app"
	"paint.pecet.it/internal/repo/env"
	"paint.pecet.it/internal/simpleauth"
	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
)

type Api struct {
	ward       *ward.Ward
	wardsocket *wardsocket.WardSocket
	auth       *simpleauth.SimpleAuth
}

const bufSize = 1024 * 64 * 4

func New(app *app.App) *Api {

	return &Api{ward: app.Ward, wardsocket: app.Wardsocket, auth: app.Auth}
}

func (api *Api) Run() {

	public := api.ward.NewGroup("/api")
	public.Post("/login", api.auth.LoginHandler)
	public.Get("/hello", api.handleHelloWorld)

	protected := api.ward.NewGroup("/api").With(api.auth.AuthMiddleware)
	protected.Get("/ws", api.handlePaintWS)
	protected.Get("/ping", api.auth.PingHandler)

	api.ward.Get("/", func(wreq *ward.Request) {
		distPath := "./dist"
		path := filepath.Join(distPath, wreq.Http.URL.Path)
		info, err := os.Stat(path)

		if os.IsNotExist(err) || info.IsDir() {
			http.ServeFile(wreq.ResponseWriter, wreq.Http, filepath.Join(distPath, "index.html"))
			return
		}

		http.ServeFile(wreq.ResponseWriter, wreq.Http, path)
	})
	log.Println("Listening on port", env.Var.Port)
	log.Fatal(http.ListenAndServe(env.Var.Port, api.ward))
}

func (api *Api) handleHelloWorld(wreq *ward.Request) {
	wreq.Write([]byte("hello world"))
}
