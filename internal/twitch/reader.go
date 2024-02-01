package twitch

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Metadata Metadata `json:"metadata"`
	Payload  Payload  `json:"payload"`
}

type Metadata struct {
	MessageID        string    `json:"message_id"`
	MessageType      string    `json:"message_type"`
	MessageTimestamp time.Time `json:"message_timestamp"`
}

type Payload struct {
	Session      Session      `json:"session"`
	Subscription Subscription `json:"subscription"`
	Event        Event        `json:"event"`
}

type Session struct {
	ID string `json:"id"`
}

type Subscription struct {
	Type string `json:"type"`
}

type Event struct {
	BroadcasterUserName string      `json:"broadcaster_user_name"`
	BroadcasterUserID   string      `json:"broadcaster_user_id"`
	ChatterUserName     string      `json:"chatter_user_name"`
	ChatMessage         ChatMessage `json:"message"`
	Color               string      `json:"color"`
	ModeratorUserLogin  string      `json:"moderator_user_login"`
	UserLogin           string      `json:"user_login"`
	IsPermanent         bool        `json:"is_permanent"`
	BannedAt            time.Time   `json:"banned_at"`
	EndsAt              time.Time   `json:"ends_at"`
	Reason              string      `json:"reason"`
}

type ChatMessage struct {
	Text string `json:"text"`
}

func Read(accessToken string, cond Condition, wsChan chan Payload, ctx context.Context) {
	defer close(wsChan)

	log.Printf("Joining %s as user %s\n", cond.BroadcasterUserID, cond.UserID)
	u := url.URL{Scheme: "wss", Host: "eventsub.wss.twitch.tv", Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer c.Close()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Parted %s\n", cond.BroadcasterUserID)
			return
		default:
			_, data, err := c.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			err = handleMsg(data, accessToken, cond, wsChan)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func handleMsg(data []byte, accessToken string, cond Condition, wsChan chan Payload) error {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}

	if msg.Metadata.MessageType == "session_welcome" {
		_, err := createEventSub(accessToken, msg.Payload.Session.ID, cond, MessageType)
		if err != nil {
			return err
		}

		_, err = createEventSub(accessToken, msg.Payload.Session.ID, cond, BanType)
		if err != nil {
			if errors.Is(err, ErrForbidden) {
				log.Printf("User %s is not mod in channel %s\n", cond.UserID, cond.BroadcasterUserID)
			} else {
				return err
			}
		}

		_, err = createEventSub(accessToken, msg.Payload.Session.ID, cond, UnbanType)
		if err != nil {
			if errors.Is(err, ErrForbidden) {
				log.Printf("User %s is not mod in channel %s\n", cond.UserID, cond.BroadcasterUserID)
			} else {
				return err
			}
		}
	}

	if msg.Metadata.MessageType == "session_reconnect" {
		// TODO: implement reconnect logic
		log.Println("session_reconnect")
	}

	if msg.Metadata.MessageType == "revocation" {
		// TODO: what do we do in this case?
		log.Println("revocation")
	}

	if msg.Metadata.MessageType == "notification" {
		wsChan <- msg.Payload
	}

	return nil
}
