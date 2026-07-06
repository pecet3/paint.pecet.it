package main

import (
	"paint.pecet.it/internal/api"
	"paint.pecet.it/internal/repo/app"
)

func main() {
	app := app.New()
	api.New(app).Run()
}
