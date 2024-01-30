package twitch

import (
	"context"
	"encoding/json"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type websocketMessage struct {
	messageType int
	data        []byte
	err         error
}

func Read(auth Authentication, cond Condition, wsChan chan Event, ctx context.Context) {
	u := url.URL{Scheme: "wss", Host: "eventsub.wss.twitch.tv", Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer c.Close()

	msgChan := make(chan websocketMessage)

	ctx, cancel := context.WithCancel(ctx)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Printf("Parted %s\n", cond.BroadcasterUserId)
				close(msgChan)
				return
			default:
				messageType, message, err := c.ReadMessage()
				msgChan <- websocketMessage{
					messageType: messageType,
					data:        message,
					err:         err,
				}
			}

		}
	}(ctx)

	for {
		wsMsg, ok := <-msgChan
		if !ok {
			cancel()
			return
		}
		err := handleMsg(wsMsg, auth, cond, wsChan)
		if err != nil {
			log.Println(err)
			cancel()
			return
		}
	}
}

func handleMsg(wsMsg websocketMessage, auth Authentication, cond Condition, wsChan chan Event) error {
	if wsMsg.err != nil {
		return wsMsg.err
	}

	var msg Message
	err := json.Unmarshal(wsMsg.data, &msg)
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

	wsChan <- msg.Payload.Event
	return nil
}
