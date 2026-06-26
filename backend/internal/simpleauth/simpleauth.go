package simpleauth

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"paint.pecet.it/pkg/ward"
)

type User struct {
	uuid string
	name string
}

func (u *User) Uuid() string {
	return u.uuid
}

func (u *User) Name() string {
	return u.name
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
	Name string `json:"name"`
}

func (sa *SimpleAuth) LoginHandler(wreq *ward.Request) {
	if wreq.Http.Method != http.MethodPost {
		http.Error(wreq.ResponseWriter, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody LoginRequest
	err := json.NewDecoder(wreq.Http.Body).Decode(&reqBody)
	if err != nil {
		http.Error(wreq.ResponseWriter, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if reqBody.Name == "" {
		http.Error(wreq.ResponseWriter, "Name is required", http.StatusBadRequest)
		return
	}

	user := &User{
		uuid: uuid.NewString(),
		name: reqBody.Name,
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
