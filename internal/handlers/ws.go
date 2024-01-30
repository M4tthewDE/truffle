package handlers

import (
	"bytes"
	"context"
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
	msgTemplate   *template.Template
	unbanTemplate *template.Template
}

func NewWsChatHandler() (*WsChatHandler, error) {
	msgTemplate, err := template.ParseFiles("resources/message.html")
	if err != nil {
		return nil, err
	}

	unbanTemplate, err := template.ParseFiles("resources/unban_message.html")
	if err != nil {
		return nil, err
	}
	return &WsChatHandler{
		msgTemplate:   msgTemplate,
		unbanTemplate: unbanTemplate,
	}, nil
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

type UnbanMessageData struct {
	CreatedAt          string
	ModeratorUserLogin string
	UserLogin          string
}

func NewUnbanMessageData(payload twitch.Payload) UnbanMessageData {
	return UnbanMessageData{
		CreatedAt:          time.Now().Format(time.TimeOnly),
		ModeratorUserLogin: payload.Event.ModeratorUserLogin,
		UserLogin:          payload.Event.UserLogin,
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
	conn := make(chan twitch.Payload)
	ctx, cancel := context.WithCancel(context.Background())
	go twitch.Read(auth, cond, conn, ctx)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		cancel()
		return
	}

	defer c.Close()

	go func(cancel context.CancelFunc) {
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				cancel()
				return
			}
		}
	}(cancel)

	for {
		payload, ok := <-conn
		if !ok {
			return
		}

		var templateBuffer bytes.Buffer
		switch payload.Subscription.Type {
		case twitch.MESSAGE_TYPE:
			if handler.msgTemplate.Execute(&templateBuffer, NewMessageData(payload.Event)); err != nil {
				log.Println(err)
				return
			}
		case twitch.UNBAN_TYPE:
			if handler.unbanTemplate.Execute(&templateBuffer, NewUnbanMessageData(payload)); err != nil {
				log.Println(err)
				return
			}
		}

		if c.WriteMessage(websocket.TextMessage, templateBuffer.Bytes()); err != nil {
			log.Println(err)
			return
		}
	}
}
