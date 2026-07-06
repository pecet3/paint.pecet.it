package app

import (
	"net/http"

	"paint.pecet.it/internal/paint"
	"paint.pecet.it/internal/repo/env"
	"paint.pecet.it/internal/simpleauth"
	"paint.pecet.it/pkg/ward"
	"paint.pecet.it/pkg/ward/wardsocket"
)

type App struct {
	Ward       *ward.Ward
	Wardsocket *wardsocket.WardSocket
	Auth       *simpleauth.SimpleAuth
	Paint      *paint.Paint
}

const bufSize = 1024 * 64 * 4

func New() *App {
	env.Init()
	app := &App{}

	w := ward.New()
	app.Ward = w
	ws := wardsocket.New(&wardsocket.Upgrader{
		ReadBufferSize:  bufSize,
		WriteBufferSize: bufSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})
	app.Wardsocket = ws
	paint := paint.New(app.Wardsocket)
	app.Paint = paint
	auth := simpleauth.New()
	app.Auth = auth
	return app
}
