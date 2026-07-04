package ward

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
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
func (r *Request) AssignUser(user User) {
	r.User = user
}

func (r *Request) LogInfo() string {
	return fmt.Sprintf("request id: %d user uuid: %s", r.Id, r.User.Uuid())
}

func (r *Request) Log(v ...any) {
	args := append([]any{r.LogInfo()}, v...)
	log.Println(args...)
}

func (r *Request) Logf(format string, v ...any) {
	prefix := r.LogInfo() + " "
	log.Printf(prefix+format, v...)
}

func (r *Request) WriteErr(status int, msg ...string) {
	if len(msg) > 0 {
		m := strings.Join(msg, " ")
		http.Error(r.ResponseWriter, m, status)
	} else {
		http.Error(r.ResponseWriter, http.StatusText(status), status)
	}
}
