package wardrouter

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator"
)

var valid = validator.New()
var mux = http.NewServeMux()

type Ward struct {
	reqCounter uint64
	mux        *http.ServeMux

	getRoutes    map[string]Route
	postRoutes   map[string]Route
	putRoutes    map[string]Route
	deleteRoutes map[string]Route

	routes map[string]Route

	basePath string
}

func New() *Ward {
	return &Ward{
		mux:    http.NewServeMux(),
		routes: make(map[string]Route),
	}
}

type Route struct {
	path        string
	middlewares Middleware
	minimalRank int
	method      string
	handlers    []Handler
}

func (ward *Ward) HandleFunc(pattern string, handler func(wreq *Request)) {
	ward.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		wreq := newRequest(ward, r)
		wreq.ResponseWriter = w
		wreq.Http = r
		wreq.Log(wreq.ClientInfo.Ip, r.Method, r.URL.Path)
		handler(wreq)
	})
}

func (ward *Ward) middleware(pattern string, handler func(wreq *Request), mws ...Middleware) {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	ward.HandleFunc(pattern, handler)
}

func (ward *Ward) Get(path string, handler func(wreq *Request)) *Route {
	r := Route{
		path:   path,
		method: "GET ",
	}
	r.handlers[0] = handler
	ward.getRoutes[r.method+path] = r
	return &r
}

func (ward *Ward) Put(pattern string, handler func(wreq *Request), mws ...Middleware) {
	ward.middleware("PUT "+pattern, handler, mws...)
}

func (ward *Ward) Post(pattern string, handler func(wreq *Request), mws ...Middleware) {
	ward.middleware("POST "+pattern, handler, mws...)
}

func (ward *Ward) Delete(pattern string, handler func(wreq *Request), mws ...Middleware) {
	ward.middleware("DELETE "+pattern, handler, mws...)
}

type Middleware func(Handler) Handler
type Handler func(wreq *Request)

func (ward *Ward) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	log.Println(r.Method)
	mux.ServeHTTP(w, r)
}

func MinimalRank(rank int) Middleware {
	return func(next Handler) Handler {
		return func(wreq *Request) {
			if wreq.User.Rank() < rank {
				wreq.WriteErrLog(errors.New("unauthorized user: "+wreq.User.Uuid()),
					http.StatusUnauthorized, "")
				return
			} else {
				next(wreq)
			}
		}
	}
}
