package handlers

import (
	"bytes"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m4tthewde/truffle/internal/config"
	"github.com/m4tthewde/truffle/internal/session"
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

type MessageData struct {
	ChatterUserName string
	Text            string
	Color           string
	CreatedAt       string
}

func NewMessageData(event twitch.Event) MessageData {
	return MessageData{
		ChatterUserName: event.ChatterUserName,
		Text:            event.ChatMessage.Text,
		Color:           event.Color,
		CreatedAt:       time.Now().Format(time.TimeOnly),
	}
}

var upgrader = websocket.Upgrader{}

// FIXME: this sometimes takes very long (10+ seconds) to connect
func (handler *WsChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s, ok, err := session.SessionFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	channels, ok := r.URL.Query()["channel"]
	if !ok || len(channels) == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	auth := twitch.NewAuthentication(config.Conf.ClientId, s.AccessToken)

	channelId, err := twitch.GetChannelId(auth, channels[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cond := twitch.NewCondition(channelId, s.UserId)
	conn := make(chan twitch.Event)
	id := twitch.Join(auth, cond, conn)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c.SetCloseHandler(func(code int, text string) error {
		twitch.Part(cond.BroadcasterUserId, id)
		return nil
	})

	defer c.Close()

	// need to read so that we can receive close messages to the callback
	go func() {
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	for {
		event, ok := <-conn
		if !ok {
			return
		}
		var templateBuffer bytes.Buffer
		if handler.msgTemplate.Execute(&templateBuffer, NewMessageData(event)); err != nil {
			log.Println(err)
			return
		}

		if c.WriteMessage(websocket.TextMessage, templateBuffer.Bytes()); err != nil {
			log.Println(err)
			return
		}
	}
}
