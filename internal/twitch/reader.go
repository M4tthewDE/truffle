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
	MessageId        string    `json:"message_id"`
	MessageType      string    `json:"message_type"`
	MessageTimestamp time.Time `json:"message_timestamp"`
}

type Payload struct {
	Session      Session      `json:"session"`
	Subscription Subscription `json:"subscription"`
	Event        Event        `json:"event"`
}

type Session struct {
	Id string `json:"id"`
}

type Subscription struct {
	Type string `json:"type"`
}

type Event struct {
	BroadcasterUserName string      `json:"broadcaster_user_name"`
	BroadcasterUserId   string      `json:"broadcaster_user_id"`
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

func Read(auth Authentication, cond Condition, wsChan chan Payload, ctx context.Context) {
	log.Printf("Joining %s as user %s\n", cond.BroadcasterUserId, cond.UserId)
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
			log.Printf("Parted %s\n", cond.BroadcasterUserId)
			close(wsChan)
			return
		default:
			_, data, err := c.ReadMessage()
			if err != nil {
				log.Println(err)
				close(wsChan)
				return
			}

			err = handleMsg(data, auth, cond, wsChan)
			if err != nil {
				log.Println(err)
				close(wsChan)
				return
			}
		}
	}
}

func handleMsg(data []byte, auth Authentication, cond Condition, wsChan chan Payload) error {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}

	if msg.Metadata.MessageType == "session_welcome" {
		_, err := createEventSub(auth, msg.Payload.Session.Id, cond, MESSAGE_TYPE)
		if err != nil {
			return err
		}

		_, err = createEventSub(auth, msg.Payload.Session.Id, cond, BAN_TYPE)
		if err != nil {
			if errors.Is(err, ForbiddenError) {
				log.Printf("User %s is not mod in channel %s\n", cond.UserId, cond.BroadcasterUserId)
			} else {
				return err
			}
		}

		_, err = createEventSub(auth, msg.Payload.Session.Id, cond, UNBAN_TYPE)
		if err != nil {
			if errors.Is(err, ForbiddenError) {
				log.Printf("User %s is not mod in channel %s\n", cond.UserId, cond.BroadcasterUserId)
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
