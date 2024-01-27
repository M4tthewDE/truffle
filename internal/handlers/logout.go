package handlers

import (
	"net/http"

	"github.com/m4tthewde/truffle/internal/util"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionId, err := util.SessionIdFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		_, loggedIn := Sessions[*sessionId]
		if !loggedIn {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	delete(Sessions, *sessionId)
	w.WriteHeader(http.StatusNoContent)
}
