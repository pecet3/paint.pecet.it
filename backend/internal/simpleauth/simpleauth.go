package simpleauth

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"paint.pecet.it/internal/repo/env"
	"paint.pecet.it/pkg/ward"
)

type User struct {
	Uuid_ string `json:"uuid"`
	Name_ string `json:"name"`
	Rank_ int    `json:"rank"`
}

func (u *User) Uuid() string {
	return u.Uuid_
}

func (u *User) Name() string {
	return u.Name_
}
func (u *User) Rank() int {
	return u.Rank_
}

type SimpleAuth struct {
	mu    sync.RWMutex
	users map[string]*User
}

func New() *SimpleAuth {
	return &SimpleAuth{
		users: make(map[string]*User),
	}
}

type LoginRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=32"`
	Password string `json:"password,omitempty"`
}

func (sa *SimpleAuth) LoginHandler(wreq *ward.Request) {
	var reqBody LoginRequest
	err := wreq.GetValidJson(&reqBody)
	if err != nil {
		wreq.WriteErrLog(err, http.StatusBadRequest)
		return
	}

	if reqBody.Name == "" {
		http.Error(wreq.ResponseWriter, "Name is required", http.StatusBadRequest)
		return
	}
	rank := 0
	if reqBody.Password == env.Var.AdminPassword {
		rank = 100
	}
	user := &User{
		Uuid_: uuid.NewString(),
		Name_: reqBody.Name,
		Rank_: rank,
	}

	authToken := uuid.NewString()
	refreshToken := uuid.NewString()

	sa.mu.Lock()
	sa.users[authToken] = user
	sa.mu.Unlock()

	farFuture := time.Now().AddDate(100, 0, 0)

	http.SetCookie(wreq.ResponseWriter, &http.Cookie{
		Name:     "auth-token",
		Value:    authToken,
		Path:     "/",
		Expires:  farFuture,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(wreq.ResponseWriter, &http.Cookie{
		Name:     "auth-refresh-token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  farFuture,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	wreq.ResponseWriter.Write([]byte("Logged successful"))
}

func (sa *SimpleAuth) PingHandler(wreq *ward.Request) {
	user := wreq.User

	response := map[string]string{
		"uuid": user.Uuid(),
		"name": user.Name(),
	}

	wreq.ResponseWriter.Header().Set("Content-Type", "application/json")
	wreq.ResponseWriter.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(wreq.ResponseWriter).Encode(response)
}

func (sa *SimpleAuth) AuthMiddleware(next ward.Handler) ward.Handler {
	return func(wreq *ward.Request) {
		log.Println(12)

		cookie, err := wreq.Http.Cookie("auth-token")
		if err != nil {
			http.Error(wreq, "", http.StatusUnauthorized)
			return
		}

		sa.mu.RLock()
		user, exists := sa.users[cookie.Value]
		sa.mu.RUnlock()

		if !exists {
			http.Error(wreq, "", http.StatusUnauthorized)
			return
		}

		wreq.AssignUser(user)

		next(wreq)
	}
}
