package main

import (
	"net/http"

	"paint.pecet.it/internal/repo/env"
	"paint.pecet.it/pkg/wardrouter"
)

func main() {
	ward := wardrouter.New()
	ward.Get("/test", func(wreq *wardrouter.Request) {
		wreq.Write([]byte("ok"))
	})
	http.ListenAndServe(env.Var.Port, ward)
}
