package handlers

import (
	"log"
	"net/http"
	"text/template"

	"github.com/m4tthewde/truffle/internal/session"
)

type ChatBoxHandler struct {
	chatBoxTemplate *template.Template
}

func NewChatBoxHandler() (*ChatBoxHandler, error) {
	chatBoxTemplate, err := template.ParseFiles("resources/chatbox.html")
	if err != nil {
		return nil, err
	}

	return &ChatBoxHandler{chatBoxTemplate: chatBoxTemplate}, nil
}

type ChatBoxData struct {
	Channel string
}

func (handler *ChatBoxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
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

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	channels, ok := r.Form["channel"]
	if !ok || len(channels) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	channel := channels[0]

	err = handler.chatBoxTemplate.Execute(w, ChatBoxData{Channel: channel})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
