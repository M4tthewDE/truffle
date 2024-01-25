package internal

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

type RootHandler struct {
	indexTemplate *template.Template
}

func NewRootHandler() (*RootHandler, error) {
	indexTemplate, err := template.ParseFiles("resources/index.html")
	if err != nil {
		return nil, err
	}

	return &RootHandler{indexTemplate: indexTemplate}, nil
}

type RootData struct {
	LoggedIn bool
	ClientId string
}

func (handler *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var loggedIn bool
	sessionId, err := sessionIdFromRequest(r)
	if err != nil {
		loggedIn = false
	} else {
		_, loggedIn = Sessions[*sessionId]
	}

	rootData := RootData{
		LoggedIn: loggedIn,
		ClientId: os.Getenv("CLIENT_ID"),
	}
	err = handler.indexTemplate.Execute(w, rootData)
	if err != nil {
		log.Fatalln(err)
	}
}

func sessionIdFromRequest(r *http.Request) (*uuid.UUID, error) {
	sessionCookie, err := r.Cookie("sessionid")
	if err != nil {
		return nil, err
	}

	splits := strings.Split(sessionCookie.String(), "=")
	if len(splits) < 2 {
		return nil, errors.New("invalid cookie format")
	}

	sessionId, err := uuid.Parse(splits[1])
	if err != nil {
		return nil, err
	}

	return &sessionId, nil
}
