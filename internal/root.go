package internal

import (
	"log"
	"net/http"
	"text/template"
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
		ClientId: Conf.ClientId,
	}
	err = handler.indexTemplate.Execute(w, rootData)
	if err != nil {
		log.Fatalln(err)
	}
}
