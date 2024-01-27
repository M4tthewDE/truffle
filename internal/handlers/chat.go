package handlers

import (
	"log"
	"net/http"
	"text/template"

	"github.com/m4tthewde/truffle/internal/session"
	"github.com/m4tthewde/truffle/internal/twitch"
)

type ChatHandler struct {
	// TODO: rename this
	settingsTemplate *template.Template
}

func NewChatHandler() (*ChatHandler, error) {
	settingsTemplate, err := template.ParseFiles("resources/chat.html")
	if err != nil {
		return nil, err
	}

	return &ChatHandler{settingsTemplate: settingsTemplate}, nil
}

type ChatData struct {
	SessionId string
}

func (handler *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	err = handler.settingsTemplate.Execute(w, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func handleEvent(event twitch.Event) error {
	for _, eventChan := range EventChans[event.BroadcasterUserId] {
		eventChan <- event
	}

	return nil
}
