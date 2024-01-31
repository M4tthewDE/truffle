package handlers

import (
	"log"
	"net/http"
	"text/template"

	"github.com/m4tthewde/truffle/internal/session"
)

type ChatRoomHandler struct {
	chatRoomTemplate *template.Template
}

func NewChatRoomHandler() (*ChatRoomHandler, error) {
	chatRoomTemplate, err := template.ParseFiles("resources/chatroom.html")
	if err != nil {
		return nil, err
	}

	return &ChatRoomHandler{chatRoomTemplate: chatRoomTemplate}, nil
}

type ChatRoomData struct {
	Channel string
}

func (handler *ChatRoomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	channels, ok := r.Form["channel"]
	if !ok || len(channels) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	channel := channels[0]

	err = handler.chatRoomTemplate.Execute(w, ChatRoomData{Channel: channel})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
