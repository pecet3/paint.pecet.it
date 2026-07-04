package env

import (
	"log"

	"github.com/plaid/go-envvar/envvar"
)

type EnvVars struct {
	Port          string `envvar:"PORT" default:":8080"`
	StaticFolder  string `envvar:"STATIC_FOLDER" default:"./dist"`
	AdminPassword string `envvar:"ADMIN_PASSWORD" default:"bob_ross"`
}

var Var EnvVars

func Init() {
	if err := envvar.Parse(&Var); err != nil {
		log.Fatal(err)
	}
}
