package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/m4tthewde/truffle/internal/config"
	"github.com/m4tthewde/truffle/internal/session"
	"github.com/m4tthewde/truffle/internal/twitch"
)

type UserInfo struct {
	AccessToken string
	Login       string
	UserID      string
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	login, err := twitch.GetToken(
		params.Get("code"),
		config.Conf.ClientID,
		config.Conf.ClientSecret,
		config.Conf.URL,
	)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	validation, err := twitch.ValidateToken(login.AccessToken)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionID := uuid.New()
	s := session.NewSession(sessionID, login.AccessToken, validation.Login, validation.UserID)
	session.AddSession(s)
	http.Redirect(w, r, fmt.Sprintf("%s/#%s", config.Conf.URL, sessionID.String()), http.StatusFound)
}
