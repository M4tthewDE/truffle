package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// FIXME: when do these get cleaned up?
	MsgChans map[string][]chan Message
)

type ChatHandler struct {
	settingsTemplate *template.Template
}

func NewChatHandler() (*ChatHandler, error) {
	settingsTemplate, err := template.ParseFiles("resources/chat.html")
	if err != nil {
		return nil, err
	}

	return &ChatHandler{settingsTemplate: settingsTemplate}, nil
}

type ChatData struct {
	SessionId string
}

func (handler *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionId, err := sessionIdFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		_, loggedIn := Sessions[*sessionId]
		if !loggedIn {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	userId := Sessions[*sessionId].UserId
	_, alreadyConnect := MsgChans[userId]
	if !alreadyConnect {
		go readChat(Sessions[*sessionId])
	}

	err = handler.settingsTemplate.Execute(w, ChatData{SessionId: sessionId.String()})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func readChat(userInfo UserInfo) {
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

		err = handleRawMessage(message, userInfo)
		if err != nil {
			log.Fatalln(err)
		}
	}
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
	ChatterUserName string      `json:"chatter_user_name"`
	ChatMessage     ChatMessage `json:"message"`
	Color           string      `json:"color"`
}

type ChatMessage struct {
	Text string `json:"text"`
}

func handleRawMessage(rawMsg []byte, userInfo UserInfo) error {
	var msg Message
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		return err
	}

	if msg.Metadata.MessageType == "session_welcome" {
		createSub(userInfo, msg.Payload.Session.Id)
	}

	if msg.Metadata.MessageType == "session_reconnect" {
		// TODO: implement reconnect logic
	}

	if msg.Metadata.MessageType == "revocation" {
		// TODO: what do we do in this case?
	}

	if msg.Metadata.MessageType == "notification" {
		for _, msgChan := range MsgChans[userInfo.UserId] {
			msgChan <- msg
		}
	}

	return nil
}

func createSub(userInfo UserInfo, sessionId string) error {
	condition := make(map[string]string)
	condition["broadcaster_user_id"] = userInfo.UserId
	condition["user_id"] = userInfo.UserId

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
	req.Header.Add("Client-Id", Conf.ClientId)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", userInfo.AccessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 202 {
		return errors.New(resp.Status)
	}

	return nil
}

type WsChatHandler struct {
	msgTemplate *template.Template
}

func NewWsChatHandler() (*WsChatHandler, error) {
	msgTemplate, err := template.ParseFiles("resources/message.html")
	if err != nil {
		return nil, err
	}
	return &WsChatHandler{msgTemplate: msgTemplate}, nil
}

var upgrader = websocket.Upgrader{}

func (handler *WsChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionId, err := sessionIdFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		_, loggedIn := Sessions[*sessionId]
		if !loggedIn {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()

	userId := Sessions[*sessionId].UserId
	msgChan := make(chan Message)
	MsgChans[userId] = append(MsgChans[userId], msgChan)

	for {
		msg := <-msgChan
		var templateBuffer bytes.Buffer
		if handler.msgTemplate.Execute(&templateBuffer, msg); err != nil {
			log.Println(err)
			return
		}

		if c.WriteMessage(websocket.TextMessage, templateBuffer.Bytes()); err != nil {
			log.Println(err)
			return
		}
	}
}
