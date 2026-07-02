package main

import (
	"paint.pecet.it/internal/api"
	"paint.pecet.it/internal/repo/app"
	"paint.pecet.it/internal/repo/env"
)

func main() {
	env.Init()
	app := app.New()
	api.New(app).Run()
}
