package twitch

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	MessageId        string    `json:"message_id"`
	MessageType      string    `json:"message_type"`
	MessageTimestamp time.Time `json:"message_timestamp"`
}

func ReadChat() {
	u := url.URL{Scheme: "wss", Host: "eventsub.wss.twitch.tv", Path: "/ws"}
	log.Println("Connecting to eventsub websocket")

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Fatalln(err)
		}

		err = handleRawMessage(message)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func handleRawMessage(rawMsg []byte) error {
	var msg Message
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		return err
	}

	log.Println(msg)
	return nil
}
