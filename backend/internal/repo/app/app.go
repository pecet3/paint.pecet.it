package app

import (
	"net/http"

	"paint.pecet.it/internal/simpleauth"
	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
)

type App struct {
	Ward       *ward.Ward
	Wardsocket *wardsocket.WardSocket
	Auth       *simpleauth.SimpleAuth
}

const bufSize = 1024 * 64 * 4

func New() *App {
	app := &App{}

	w := ward.New()
	app.Ward = w
	ws := wardsocket.New(w, &wardsocket.Upgrader{
		ReadBufferSize:  bufSize,
		WriteBufferSize: bufSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})
	app.Wardsocket = ws
	auth := simpleauth.New()
	app.Auth = auth
	return app
}
