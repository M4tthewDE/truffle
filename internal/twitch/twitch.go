package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

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

func CreateMessageSub(authentication Authentication, sessionId string, condition Condition) error {
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
	ChatterUserName     string      `json:"chatter_user_name"`
	ChatMessage         ChatMessage `json:"message"`
	Color               string      `json:"color"`
}

type ChatMessage struct {
	Text string `json:"text"`
}
