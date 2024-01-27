package handlers

import (
	"net/http"

	"github.com/m4tthewde/truffle/internal/session"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	s, ok, err := session.SessionFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	session.DeleteSession(s)
	w.WriteHeader(http.StatusNoContent)
}
