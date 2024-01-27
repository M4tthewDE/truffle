package internal

import (
	"bytes"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/websocket"
	"github.com/m4tthewde/truffle/internal/twitch"
)

type WsChatHandler struct {
	msgTemplate *template.Template
}

func NewWsChatHandler() (*WsChatHandler, error) {
	msgTemplate, err := template.ParseFiles("resources/message.html")
	if err != nil {
		return nil, err
	}
	return &WsChatHandler{msgTemplate: msgTemplate}, nil
}

var upgrader = websocket.Upgrader{}

// FIXME: this sometimes takes very long (10+ seconds) to connect
func (handler *WsChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionId, err := sessionIdFromRequest(r)
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

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()

	userId := Sessions[*sessionId].UserId
	eventChan := make(chan twitch.Event)
	EventChans[userId] = append(EventChans[userId], eventChan)

	for {
		event := <-eventChan
		var templateBuffer bytes.Buffer
		if handler.msgTemplate.Execute(&templateBuffer, event); err != nil {
			log.Println(err)
			return
		}

		if c.WriteMessage(websocket.TextMessage, templateBuffer.Bytes()); err != nil {
			log.Println(err)
			return
		}
	}
}
