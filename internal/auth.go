package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type UserInfo struct {
	AccessToken string
	Login       string
	UserId      string
}

var (
	// TODO: expire sessions after a week (like cookies)
	Sessions map[uuid.UUID]UserInfo
)

type LoginHandler struct {
}

func NewLoginHandler() (*LoginHandler, error) {
	return &LoginHandler{}, nil
}

func (handler *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	login, err := getToken(params.Get("code"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	validation, err := validateToken(login.AccessToken)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionId := uuid.New()
	Sessions[sessionId] = UserInfo{
		AccessToken: login.AccessToken,
		Login:       validation.Login,
		UserId:      validation.UserId,
	}
	http.Redirect(w, r, "http://localhost:8080/#"+sessionId.String(), http.StatusFound)
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

func getToken(code string) (*TokenResponse, error) {
	body := make(map[string]string)
	body["client_id"] = Conf.ClientId
	body["client_secret"] = Conf.ClientSecret
	body["code"] = code
	body["grant_type"] = "authorization_code"
	body["redirect_uri"] = "http://localhost:8080/login"

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

func validateToken(accessToken string) (*ValidationResponse, error) {
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
