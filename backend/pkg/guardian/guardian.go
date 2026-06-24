package guardian

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

type Guardian struct {
	reqCounter  uint64
	mux         *http.ServeMux
	httpClients map[string]*ClientInfo
	hMu         sync.RWMutex
}

func New() *Guardian {
	return &Guardian{
		mux:         http.NewServeMux(),
		httpClients: make(map[string]*ClientInfo),
	}
}

func (g *Guardian) HandleFunc(pattern string, handler func(w http.ResponseWriter, req *http.Request)) {
	g.mux.HandleFunc(pattern, handler)
}

func (g *Guardian) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := GetClientIP(r)
	g.hMu.RLock()
	client, ok := g.httpClients[ip]
	g.hMu.RUnlock()
	if !ok {
		g.hMu.Lock()
		if client, ok = g.httpClients[ip]; !ok {
			client = &ClientInfo{
				Ip: ip,
			}
			g.httpClients[ip] = client
		}
		g.hMu.Unlock()
	}
	gReq := &Request{
		Id:         atomic.AddUint64(&g.reqCounter, 1),
		ClientInfo: client,
		User:       &TestUser{},
		Req:        r,
		W:          w,
	}

	rWithContext := SetGuardianRequest(r, gReq)

	g.mux.ServeHTTP(w, rWithContext)
}

type TestUser struct {
}

func (u *TestUser) Uuid() string {
	return uuid.NewString()
}
func (u *TestUser) Name() string {
	return "test user"
}
