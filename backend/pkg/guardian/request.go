package guardian

import (
	"context"
	"net/http"
)

type User interface {
	Name() string
	Uuid() string
}

type Request struct {
	ClientInfo *ClientInfo
	User       User
	Req        *http.Request
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
