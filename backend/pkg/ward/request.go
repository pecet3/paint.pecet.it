package ward

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/goccy/go-json"
)

type User interface {
	Name() string
	Uuid() string
	Rank() int
}

type nullUser struct{}

var nUser = &nullUser{}

const (
	null = "null"
)

func (u *nullUser) Uuid() string {
	return null
}
func (u *nullUser) Name() string {
	return null
}
func (u *nullUser) Rank() int {
	return 0
}

type ClientInfo struct {
	Ip string
}

type Request struct {
	http.ResponseWriter
	Id         uint64
	Http       *http.Request
	ClientInfo *ClientInfo
	User       User
}

func newRequest(ward *Ward, r *http.Request) *Request {
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
		Id:         atomic.AddUint64(&ward.reqCounter, 1),
		ClientInfo: client,
		User:       nUser,
	}

	return wreq
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
	r.Log("UserUUID:", r.User.Uuid(), "UserName:", r.User.Name())
}

func (r *Request) LogInfo() string {
	return fmt.Sprintf("Reqest %d |", r.Id)
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

func (r *Request) WriteErrLog(err error, status int, msg ...string) {
	r.Log(err)
	r.WriteErr(status, msg...)
}

func (r *Request) WriteJson(v any) error {
	return json.NewEncoder(r.ResponseWriter).Encode(v)
}

func (r *Request) WriteJsonOrErrLog(v any, errMsg ...string) {
	err := r.WriteJson(v)
	if err != nil {
		r.WriteErrLog(err, http.StatusInternalServerError, errMsg...)
	}
}

func (r *Request) GetJson(v any) error {
	return json.NewDecoder(r.Http.Body).Decode(v)
}

func (r *Request) GetValidJson(v any) error {
	err := r.GetJson(v)
	if err != nil {
		return err
	}
	err = valid.Struct(v)

	return err
}
