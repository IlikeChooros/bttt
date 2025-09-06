package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var CookieStore *sessions.CookieStore = nil

func InitAuth() {
	// In production, use a secure random key and keep it secret
	CookieStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
}

func Authenticate(w http.ResponseWriter, r *http.Request) (string, error) {
	// Use session cookies or headers to authenticate the user
	session, err := CookieStore.Get(r, "session-name")
	if err != nil {
		return "", err
	}

	if session.IsNew {
		session.Values["userId"] = RandID(16)
		err = session.Save(r, w)
		if err != nil {
			return "", err
		}
	}

	userId, ok := session.Values["userId"].(string)
	if !ok {
		return "", fmt.Errorf("userId not found in session")
	}
	return userId, nil
}
