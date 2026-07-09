package api

import (
	"log"
	"net/http"

	"paint.pecet.it/internal/paint"
	"paint.pecet.it/internal/repo/app"
	"paint.pecet.it/internal/repo/env"
	"paint.pecet.it/internal/simpleauth"
	"paint.pecet.it/pkg/ward"
)

type Api struct {
	ward  *ward.Ward
	auth  *simpleauth.SimpleAuth
	paint *paint.Paint
}

func New(app *app.App) *Api {
	app.Paint.CreateRoom(&paint.RoomConfig{
		Name:        "general",
		IsTemporary: false,
	})
	return &Api{ward: app.Ward, auth: app.Auth, paint: app.Paint}
}

func (api *Api) registerHandlers() {
	public := api.ward.NewGroup("/api")
	public.Post("/login", api.auth.LoginHandler)
	authorized := api.ward.NewGroup("/api").Use(api.auth.AuthMiddleware)
	authorized.Get("/rooms/{id}", api.handleJoinRoom)
	authorized.Post("/rooms", api.handleCreateRoom)
	authorized.Get("/rooms", api.handleRoomsList)

	authorized.Get("/ping", api.auth.PingHandler)

	test := api.ward.NewGroup("/api").Use()
	test.Get("/test", api.test, ward.MinimalRank(0))

	api.ward.Get("/", api.handleStaticFiles)
}
func (api *Api) Run() {
	api.registerHandlers()
	log.Println("Listening on port", env.Var.Port)
	log.Fatal(http.ListenAndServe(env.Var.Port, api.ward))
}

func (Api) test(wreq *ward.Request) {
	wreq.Write([]byte("<h1>test</h1>"))
}
