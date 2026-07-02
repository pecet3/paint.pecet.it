package ward

import (
	"net/http"
	"sync"
	"sync/atomic"
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

func (g *Group) Use(mws ...Middleware) {
	g.middlewares = append(g.middlewares, mws...)
}

func (g *Group) With(mws ...Middleware) *Group {
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

func (ward *Ward) Handle(pattern string, handler func(wreq *Request)) {
	ward.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		wreq := GetWardRequest(r)
		handler(wreq)
	})
}

func (ward *Ward) middleware(pattern string, handler func(wreq *Request), mws ...Middleware) {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	ward.Handle(pattern, handler)
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
	ip := getClientIP(r)
	ward.hMu.RLock()
	client, ok := ward.httpClients[ip]
	ward.hMu.RUnlock()
	if !ok {
		ward.hMu.Lock()
		if client, ok = ward.httpClients[ip]; !ok {
			client = &ClientInfo{
				Ip: ip,
			}
			ward.httpClients[ip] = client
		}
		ward.hMu.Unlock()
	}
	wreq := &Request{
		Id:             atomic.AddUint64(&ward.reqCounter, 1),
		ClientInfo:     client,
		User:           nUser,
		Http:           r,
		ResponseWriter: w,
	}

	rWithContext := SetWardRequest(r, wreq)

	ward.mux.ServeHTTP(w, rWithContext)
}

var nUser = &nullUser{}

const (
	nullUserUuid = "null-user-uuid"
	nullUserName = "nullUserName"
)

type nullUser struct{}

func (u *nullUser) Uuid() string {
	return nullUserUuid
}

func (u *nullUser) Name() string {
	return nullUserName
}
