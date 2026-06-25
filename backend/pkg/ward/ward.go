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
}

func New() *Ward {
	return &Ward{
		mux:         http.NewServeMux(),
		httpClients: make(map[string]*ClientInfo),
	}
}

func (ward *Ward) HandleFunc(pattern string, handler func(w http.ResponseWriter, req *http.Request)) {
	ward.mux.HandleFunc(pattern, handler)
}

func (ward *Ward) Handle(pattern string, handler func(wreq *Request)) {
	ward.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		wreq := GetWardRequest(r)
		wreq.Http = r
		wreq.ResponseWriter = w
		handler(wreq)
	})
}

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
	nullUserUuid = "nullUserUuid"
	nullUserName = "nullUserName"
)

type nullUser struct {
}

func (u *nullUser) Uuid() string {
	return nullUserUuid
}
func (u *nullUser) Name() string {
	return nullUserName
}
