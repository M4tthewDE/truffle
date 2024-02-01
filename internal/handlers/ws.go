package handlers

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/gorilla/websocket"
	"github.com/m4tthewde/truffle/internal/components"
	"github.com/m4tthewde/truffle/internal/session"
	"github.com/m4tthewde/truffle/internal/twitch"
)

var upgrader = websocket.Upgrader{}

// FIXME: this sometimes takes very long (10+ seconds) to connect
func WsChatHandler(w http.ResponseWriter, r *http.Request) {
	s, ok, err := session.SessionFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channelId, err := twitch.GetChannelId(ctx, s.AccessToken, r.FormValue("channel"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	conn := make(chan twitch.Payload)
	go twitch.Read(s.AccessToken, twitch.NewCondition(channelId, s.UserId), conn, ctx)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer c.Close()

	if sendConnectMessage(ctx, c); err != nil {
		log.Println(err)
		return
	}

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
			log.Println("Reader closed connection")
			return
		}

		var templateBuffer bytes.Buffer
		switch payload.Subscription.Type {
		case twitch.MESSAGE_TYPE:
			component := components.Message(
				time.Now(),
				templ.Attributes{"style": "color:" + payload.Event.Color},
				payload.Event.ChatterUserName,
				payload.Event.ChatMessage.Text,
			)
			if component.Render(ctx, &templateBuffer); err != nil {
				log.Println(err)
				return
			}
		case twitch.UNBAN_TYPE:
			component := components.UnbanMessage(
				time.Now(),
				payload.Event.ModeratorUserLogin,
				payload.Event.UserLogin,
			)

			if component.Render(ctx, &templateBuffer); err != nil {
				log.Println(err)
				return
			}

		case twitch.BAN_TYPE:
			component := components.BanMessage(
				payload.Event.BannedAt,
				payload.Event.IsPermanent,
				payload.Event.ModeratorUserLogin,
				payload.Event.UserLogin,
				payload.Event.Reason,
				payload.Event.EndsAt.Sub(payload.Event.BannedAt),
			)
			if component.Render(ctx, &templateBuffer); err != nil {
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

func sendConnectMessage(ctx context.Context, c *websocket.Conn) error {
	var connectBuffer bytes.Buffer
	component := components.ConnectMessage()
	err := component.Render(ctx, &connectBuffer)
	if err != nil {
		return err
	}

	err = c.WriteMessage(websocket.TextMessage, connectBuffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}
