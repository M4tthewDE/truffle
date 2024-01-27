package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	Session Session `json:"session"`
	Event   Event   `json:"event"`
}

type Session struct {
	Id string `json:"id"`
}

type Event struct {
	BroadcasterUserName string      `json:"broadcaster_user_name"`
	BroadcasterUserId   string      `json:"broadcaster_user_id"`
	ChatterUserName     string      `json:"chatter_user_name"`
	ChatMessage         ChatMessage `json:"message"`
	Color               string      `json:"color"`
}

type ChatMessage struct {
	Text string `json:"text"`
}

func ReadChat(auth Authentication, condition Condition, fn func(Event) error) {
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

		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Fatalln(err)
		}

		if msg.Metadata.MessageType == "session_welcome" {
			err = createMessageSub(auth, msg.Payload.Session.Id, condition)
			if err != nil {
				log.Println(err)
			}
		}

		if msg.Metadata.MessageType == "session_reconnect" {
			// TODO: implement reconnect logic
		}

		if msg.Metadata.MessageType == "revocation" {
			// TODO: what do we do in this case?
		}

		err = fn(msg.Payload.Event)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

type Condition struct {
	BroadcasterUserId string `json:"broadcaster_user_id"`
	UserId            string `json:"user_id"`
}

func NewCondition(broadcasterUserId string, userId string) Condition {
	return Condition{
		BroadcasterUserId: broadcasterUserId,
		UserId:            userId,
	}
}

type Authentication struct {
	ClientId    string
	AccessToken string
}

func NewAuthentication(clientId string, accessToken string) Authentication {
	return Authentication{
		ClientId:    clientId,
		AccessToken: accessToken,
	}
}

func createMessageSub(authentication Authentication, sessionId string, condition Condition) error {
	transport := make(map[string]string)
	transport["method"] = "websocket"
	transport["session_id"] = sessionId

	body := make(map[string]interface{})
	body["type"] = "channel.chat.message"
	body["version"] = "1"
	body["condition"] = condition
	body["transport"] = transport

	jsonStr, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Client-Id", authentication.ClientId)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.AccessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 202 {
		return errors.New(resp.Status)
	}

	return nil
}

type ChannelResponse struct {
	Data []ChannelData `json:"data"`
}

type ChannelData struct {
	Id string `json:"id"`
}

func GetChannelId(authentication Authentication, channel string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("login", channel)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Client-Id", authentication.ClientId)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.AccessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	var channelResponse ChannelResponse
	err = json.NewDecoder(resp.Body).Decode(&channelResponse)
	if err != nil {
		return "", err
	}

	log.Println(channelResponse)

	return channelResponse.Data[0].Id, nil
}
