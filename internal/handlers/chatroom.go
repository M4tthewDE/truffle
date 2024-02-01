package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/m4tthewde/truffle/internal/components"
	"github.com/m4tthewde/truffle/internal/session"
)

func ChatRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, ok, err := session.SessionFromRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ok {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	channel := r.FormValue("channel")
	component := components.ChatRoom(
		channel,
		templ.Attributes{"ws-connect": "/chat/messages?channel=" + channel},
	)

	err = component.Render(context.Background(), w)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
