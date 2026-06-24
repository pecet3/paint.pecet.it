package ward

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type ClientInfo struct {
	LastReqDuration   time.Duration
	LastURL           string
	LastMethod        string
	LastPath          string
	ActiveConnections int
	Ip                string
}

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
	gReq := &Request{
		Id:             atomic.AddUint64(&ward.reqCounter, 1),
		ClientInfo:     client,
		User:           &TestUser{},
		Http:           r,
		ResponseWriter: w,
	}

	rWithContext := SetWardRequest(r, gReq)

	ward.mux.ServeHTTP(w, rWithContext)
}

type TestUser struct {
}

func (u *TestUser) Uuid() string {
	return uuid.NewString()
}
func (u *TestUser) Name() string {
	return "test user"
}
