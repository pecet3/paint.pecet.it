package ward

import (
	"log"
	"net/http"
	"sync"
)

type Ward struct {
	reqCounter  uint64
	mux         *http.ServeMux
	httpClients map[string]*ClientInfo
	hMu         sync.RWMutex

	basePath string
}

type Group struct {
	ward        *Ward
	path        string
	middlewares []Middleware
}

func (g *Group) Use(mws ...Middleware) *Group {
	g.middlewares = append(g.middlewares, mws...)
	return g
}

func (g *Group) combineMiddleware(mws []Middleware) []Middleware {
	combined := make([]Middleware, 0, len(g.middlewares)+len(mws))
	combined = append(combined, g.middlewares...)
	combined = append(combined, mws...)
	return combined
}
func (g *Group) NewGroup(basePath string) *Group {
	return &Group{
		ward: g.ward,
		path: g.path + basePath,
	}
}
func (g *Group) Get(pattern string, handler func(wreq *Request), mws ...Middleware) {
	g.ward.Get(g.path+pattern, handler, g.combineMiddleware(mws)...)
}

func (g *Group) Put(pattern string, handler func(wreq *Request), mws ...Middleware) {
	g.ward.Put(g.path+pattern, handler, g.combineMiddleware(mws)...)
}

func (g *Group) Post(pattern string, handler func(wreq *Request), mws ...Middleware) {
	g.ward.Post(g.path+pattern, handler, g.combineMiddleware(mws)...)
}

func (g *Group) Delete(pattern string, handler func(wreq *Request), mws ...Middleware) {
	g.ward.Delete(g.path+pattern, handler, g.combineMiddleware(mws)...)
}

func New() *Ward {
	return &Ward{
		mux:         http.NewServeMux(),
		httpClients: make(map[string]*ClientInfo),
	}
}

func (w *Ward) NewGroup(basePath string) *Group {
	return &Group{
		ward: w,
		path: w.basePath + basePath,
	}
}

func (ward *Ward) HandleFunc(pattern string, handler func(wreq *Request)) {
	ward.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		wreq := GetWardRequest(r)
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

func (ward *Ward) Get(pattern string, handler func(wreq *Request), mws ...Middleware) {
	ward.middleware("GET "+pattern, handler, mws...)
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
	log.Println(1)
	wreq := newRequest(ward, r)
	rCtx := SetWardRequest(r, wreq)
	ward.mux.ServeHTTP(w, rCtx)
}
