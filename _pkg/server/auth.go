package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

var CookieStore *sessions.CookieStore = nil

func InitAuth() {
	// In production, use a secure random key and keep it secret
	CookieStore = sessions.NewCookieStore([]byte("super-secret-key"))
}

func Authenticate(r *http.Request) (string, error) {
	// Use session cookies or headers to authenticate the user
	session, err := CookieStore.Get(r, "session-name")
	if err != nil {
		return "", err
	}
	userId, ok := session.Values["userId"].(string)
	if !ok {
		return "", fmt.Errorf("userId not found in session")
	}
	return userId, nil
}
