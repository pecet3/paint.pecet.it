package guardian

import (
	"log"
	"net/http"
	"sync"
	"time"
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
	}
	log.Println(client)
	gReq := &Request{
		ClientInfo: &ClientInfo{},
		User:       nil,
		Req:        r,
	}

	rWithContext := SetGuardianRequest(r, gReq)

	g.mux.ServeHTTP(w, rWithContext)
}
