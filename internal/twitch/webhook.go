package twitch

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
)

var (
	eventChannels map[string][]chan Event
	secret        = uuid.New()
	token         string
)

func InitEventChans(clientId string, clientSecret string) error {
	token, err := getAppAccessToken(clientId, clientSecret)
	if err != nil {
		return err
	}

	err = deleteSubscriptions(NewAuthentication(clientId, token))
	if err != nil {
		return err
	}

	eventChannels = make(map[string][]chan Event)
	return nil
}

func ListenToChannel(clientId string, cond Condition, callback string, eventChannel chan Event) error {
	channels, ok := eventChannels[cond.BroadcasterUserId]
	if !ok {
		transport := NewTransport(callback, secret.String())
		err := createMessageSub(NewAuthentication(clientId, token), cond, transport)
		if err != nil {
			return err
		}
	}

	channels = append(channels, eventChannel)
	return nil
}

type eventSubNotification struct {
	Challenge string `json:"challenge"`
	Event     Event  `json:"event"`
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// verify that the notification came from twitch using the secret
	if !verifyNotification(secret.String(), r.Header, string(body)) {
		log.Println("No valid signature on subscription!")
		return
	} else {
		log.Println("Verified signature for subscription.")
	}

	var noti eventSubNotification
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&noti)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// if there's a challenge in the request,
	// respond with only the challenge to verify your eventsub
	if noti.Challenge != "" {
		log.Println("Challenge received from twitch. Answering...")
		fmt.Fprintf(w, noti.Challenge)
		return
	}

	for _, eventChan := range eventChannels[noti.Event.BroadcasterUserId] {
		eventChan <- noti.Event
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

// from https://github.com/nicklaw5/helix/blob/7714798d932f8f7370e22651051d2801f5a02804/eventsub.go#L736
func verifyNotification(secret string, header http.Header, message string) bool {
	hmacMessage := []byte(fmt.Sprintf("%s%s%s", header.Get("Twitch-Eventsub-Message-Id"), header.Get("Twitch-Eventsub-Message-Timestamp"), message))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(hmacMessage)
	hmacsha256 := fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
	return hmacsha256 == header.Get("Twitch-Eventsub-Message-Signature")
}
