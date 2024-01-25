package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/google/uuid"
)

var (
	// TODO: expire sessions after a week (like cookies)
	Sessions map[uuid.UUID]int
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

	login, err := getLogin(params.Get("code"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := getUser(login.AccessToken)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionId := uuid.New()
	Sessions[sessionId] = user.Id
	http.Redirect(w, r, "http://localhost:8080/#"+sessionId.String(), http.StatusFound)
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

func getLogin(code string) (*LoginResponse, error) {
	body := make(map[string]string)
	body["client_id"] = os.Getenv("CLIENT_ID")
	body["client_secret"] = os.Getenv("CLIENT_SECRET")
	body["code"] = code

	jsonStr, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(jsonStr))
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

	var loginResponse LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		return nil, err
	}

	return &loginResponse, nil
}

type UserResponse struct {
	Login string `json:"login"`
	Id    int    `json:"id"`
}

func getUser(accessToken string) (*UserResponse, error) {
	body := make(map[string]string)
	body["access_token"] = accessToken

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", "https://api.github.com/user", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	var userInfo UserResponse
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}
