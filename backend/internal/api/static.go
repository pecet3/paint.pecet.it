package api

import (
	"net/http"
	"os"
	"path/filepath"

	"paint.pecet.it/internal/repo/env"
	"paint.pecet.it/pkg/ward"
)

func (api *Api) handleStaticFiles(wreq *ward.Request) {
	path := filepath.Join(env.Var.StaticFolder, wreq.Http.URL.Path)
	info, err := os.Stat(path)

	if os.IsNotExist(err) || info.IsDir() {
		http.ServeFile(wreq.ResponseWriter, wreq.Http, filepath.Join(env.Var.StaticFolder, "index.html"))
		return
	}

	http.ServeFile(wreq.ResponseWriter, wreq.Http, path)
}
