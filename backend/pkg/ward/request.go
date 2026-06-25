package ward

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type User interface {
	Name() string
	Uuid() string
}
type ClientInfo struct {
	Uuid              string
	LastReqDuration   time.Duration
	LastURL           string
	LastMethod        string
	LastPath          string
	ActiveConnections int
	Ip                string
}

type Request struct {
	http.ResponseWriter
	Id         uint64
	Http       *http.Request
	ClientInfo *ClientInfo
	User       User
}

func (r *Request) LogInfo() string {
	return fmt.Sprintln("request id:", r.Id, "user uuid:", r.User.Uuid())
}

func (r *Request) Log(v ...any) {
	log.Println(r.LogInfo(), v)
}

type contextKey string

const WardRequestKey contextKey = "WardKey"

func SetWardRequest(req *http.Request, gReq *Request) *http.Request {
	ctx := context.WithValue(req.Context(), WardRequestKey, gReq)
	return req.WithContext(ctx)
}

func GetWardRequest(req *http.Request) *Request {
	if val, ok := req.Context().Value(WardRequestKey).(*Request); ok && val != nil {
		return val
	}
	return &Request{
		ClientInfo: &ClientInfo{},
		Http:       req,
	}
}
