package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
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

type Transport struct {
	Method   string `json:"method"`
	Callback string `json:"callback"`
	Secret   string `json:"secret"`
}

func NewTransport(callback string, secret string) Transport {
	return Transport{
		Method:   "webhook",
		Callback: callback,
		Secret:   secret,
	}
}

func createMessageSub(authentication Authentication, condition Condition, transport Transport) error {
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

type SubscriptionResponse struct {
	Data []SubscriptionData `json:"data"`
}

type SubscriptionData struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

func getSubscriptions(auth Authentication) (*SubscriptionResponse, error) {
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/eventsub/subscriptions", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Client-Id", auth.ClientId)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	var subscriptionResponse SubscriptionResponse
	err = json.NewDecoder(resp.Body).Decode(&subscriptionResponse)
	if err != nil {
		return nil, err
	}

	return &subscriptionResponse, nil
}

func deleteSubscriptions(auth Authentication) error {
	subs, err := getSubscriptions(auth)
	if err != nil {
		return err
	}

	for _, subs := range subs.Data {
		log.Printf("Deleting %s with status %s", subs.Id, subs.Status)
		req, err := http.NewRequest("DELETE", "https://api.twitch.tv/helix/eventsub/subscriptions", nil)
		if err != nil {
			return err
		}
		req.Header.Add("Client-Id", auth.ClientId)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken))

		q := req.URL.Query()
		q.Add("id", subs.Id)
		req.URL.RawQuery = q.Encode()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != 204 {
			return errors.New(resp.Status)
		}

	}

	return nil
}

func getAppAccessToken(clientId string, clientSecret string) (string, error) {
	body := make(map[string]string)
	body["client_id"] = clientId
	body["client_secret"] = clientSecret
	body["grant_type"] = "client_credentials"

	jsonStr, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	var tokenResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
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

	return channelResponse.Data[0].Id, nil
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

func GetToken(code string, clientId string, clientSecret string, uri string) (*TokenResponse, error) {
	body := make(map[string]string)
	body["client_id"] = clientId
	body["client_secret"] = clientSecret
	body["code"] = code
	body["grant_type"] = "authorization_code"
	body["redirect_uri"] = fmt.Sprintf("%s/login", uri)

	jsonStr, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	var loginResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		return nil, err
	}

	return &loginResponse, nil
}

type ValidationResponse struct {
	UserId string `json:"user_id"`
	Login  string `json:"login"`
}

func ValidateToken(accessToken string) (*ValidationResponse, error) {
	req, err := http.NewRequest("GET", "https://id.twitch.tv/oauth2/validate", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", accessToken))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	var userInfo ValidationResponse
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}
