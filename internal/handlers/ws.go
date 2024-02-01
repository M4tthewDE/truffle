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

// WsChatHandler FIXME: this sometimes takes very long (10+ seconds) to connect
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

	channelID, err := twitch.GetChannelID(ctx, s.AccessToken, r.FormValue("channel"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	conn := make(chan twitch.Payload)
	go twitch.Read(s.AccessToken, twitch.NewCondition(channelID, s.UserID), conn, ctx)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer c.Close()

	var connectBuffer bytes.Buffer
	component := components.ConnectMessage()
	err = component.Render(ctx, &connectBuffer)
	if err != nil {
		log.Println(err)
		return
	}

	err = c.WriteMessage(websocket.TextMessage, connectBuffer.Bytes())
	if err != nil {
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

	// if we don't send a ping, htmx reconnects for no reason
	// htmx uses 100s as interval
	pingTicker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-pingTicker.C:
			if c.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)
				return
			}

		case payload, ok := <-conn:
			if !ok {
				log.Println("Reader closed connection")
				return
			}

			var templateBuffer bytes.Buffer
			switch payload.Subscription.Type {
			case twitch.MessageType:
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
			case twitch.UnbanType:
				component := components.UnbanMessage(
					time.Now(),
					payload.Event.ModeratorUserLogin,
					payload.Event.UserLogin,
				)

				if component.Render(ctx, &templateBuffer); err != nil {
					log.Println(err)
					return
				}

			case twitch.BanType:
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
}
