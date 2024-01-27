package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"text/template"

	"github.com/gorilla/websocket"
	"github.com/m4tthewde/truffle/internal/twitch"
)

var (
	// FIXME: when do these get cleaned up?
	MsgChans map[string][]chan twitch.Message
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

func handleRawMessage(rawMsg []byte, userInfo UserInfo) error {
	var msg twitch.Message
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		return err
	}

	if msg.Metadata.MessageType == "session_welcome" {
		err = twitch.CreateMessageSub(
			twitch.NewAuthentication(Conf.ClientId, userInfo.AccessToken),
			msg.Payload.Session.Id,
			twitch.NewCondition(userInfo.UserId, userInfo.UserId),
		)
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

	if msg.Metadata.MessageType == "notification" {
		for _, msgChan := range MsgChans[userInfo.UserId] {
			msgChan <- msg
		}
	}

	return nil
}
