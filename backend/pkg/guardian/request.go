package guardian

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type User interface {
	Name() string
	Uuid() string
}

type Request struct {
	Id  uint64
	Req *http.Request
	W   http.ResponseWriter

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

const guardianRequestKey contextKey = "guardianKey"

func SetGuardianRequest(req *http.Request, gReq *Request) *http.Request {
	ctx := context.WithValue(req.Context(), guardianRequestKey, gReq)
	return req.WithContext(ctx)
}

func GetGuardianRequest(req *http.Request) *Request {
	if val, ok := req.Context().Value(guardianRequestKey).(*Request); ok && val != nil {
		return val
	}
	return &Request{
		ClientInfo: &ClientInfo{},
		Req:        req,
	}
}
