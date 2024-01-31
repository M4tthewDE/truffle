package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/m4tthewde/truffle/internal/config"
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

type EventsubResponse struct {
	Data []EventsubData `json:"data"`
}

type EventsubData struct {
	Id string `json:"id"`
}

const (
	MESSAGE_TYPE = "channel.chat.message"
	BAN_TYPE     = "channel.ban"
	UNBAN_TYPE   = "channel.unban"
)

var ForbiddenError = errors.New("403 Forbidden")

func createEventSub(accessToken string, sessionId string, condition Condition, subType string) (string, error) {
	transport := make(map[string]string)
	transport["method"] = "websocket"
	transport["session_id"] = sessionId

	body := make(map[string]interface{})
	body["type"] = subType
	body["version"] = "1"
	body["condition"] = condition
	body["transport"] = transport

	jsonStr, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Client-Id", config.Conf.ClientId)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == 403 {
		return "", ForbiddenError
	}

	if resp.StatusCode != 202 {
		return "", errors.New(resp.Status)
	}

	var eventsubResponse EventsubResponse
	err = json.NewDecoder(resp.Body).Decode(&eventsubResponse)
	if err != nil {
		return "", err
	}

	return eventsubResponse.Data[0].Id, nil
}

func deleteMessageSub(accessToken string, id string) error {
	req, err := http.NewRequest("DELETE", "https://api.twitch.tv/helix/eventsub/subscriptions", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("id", id)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Client-Id", config.Conf.ClientId)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
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

func GetChannelId(ctx context.Context, accessToken string, channel string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("login", channel)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Client-Id", config.Conf.ClientId)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

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

func RevokeToken(ctx context.Context, accessToken string) error {
	data := url.Values{}
	data.Set("client_id", config.Conf.ClientId)
	data.Set("token", accessToken)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://id.twitch.tv/oauth2/revoke", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	return nil
}
