package handlers

import (
	"log"
	"net/http"
	"text/template"

	"github.com/m4tthewde/truffle/internal/config"
	"github.com/m4tthewde/truffle/internal/twitch"
	"github.com/m4tthewde/truffle/internal/util"
)

var (
	// FIXME: when do these get cleaned up?
	EventChans map[string][]chan twitch.Event
)

type ChatHandler struct {
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

	userId := Sessions[*sessionId].UserId
	_, alreadyConnect := EventChans[userId]
	if !alreadyConnect {
		userInfo := Sessions[*sessionId]
		auth := twitch.NewAuthentication(config.Conf.ClientId, userInfo.AccessToken)
		cond := twitch.NewCondition(userInfo.UserId, userInfo.UserId)
		go twitch.ReadChat(auth, cond, handleEvent)
	}

	err = handler.settingsTemplate.Execute(w, ChatData{SessionId: sessionId.String()})
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
