package twitch

import (
	"context"
	"encoding/json"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

func Read(auth Authentication, cond Condition, wsChan chan Event, ctx context.Context) {
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

			handleMsg(data, auth, cond, wsChan)
		}

	}
}

func handleMsg(data []byte, auth Authentication, cond Condition, wsChan chan Event) error {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}

	if msg.Metadata.MessageType == "session_welcome" {
		_, err := createMessageSub(auth, msg.Payload.Session.Id, cond)
		if err != nil {
			return err
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
		wsChan <- msg.Payload.Event
	}

	return nil
}
