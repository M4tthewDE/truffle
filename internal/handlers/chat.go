package handlers

import (
	"log"
	"net/http"

	"github.com/m4tthewde/truffle/internal/components"
	"github.com/m4tthewde/truffle/internal/session"
)

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, ok, err := session.SessionFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	component := components.Chat()

	err = component.Render(r.Context(), w)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
