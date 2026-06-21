package guardian

import (
	"context"
	"net/http"
)

type Request struct {
	ClientInfo *ClientInfo
	AuthUser   any
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
