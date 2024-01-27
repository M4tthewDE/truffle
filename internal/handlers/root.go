package handlers

import (
	"log"
	"net/http"
	"text/template"

	"github.com/m4tthewde/truffle/internal/config"
	"github.com/m4tthewde/truffle/internal/util"
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
	sessionId, err := util.SessionIdFromRequest(r)
	if err != nil {
		loggedIn = false
	} else {
		_, loggedIn = Sessions[*sessionId]
	}

	rootData := RootData{
		LoggedIn: loggedIn,
		ClientId: config.Conf.ClientId,
	}
	err = handler.indexTemplate.Execute(w, rootData)
	if err != nil {
		log.Fatalln(err)
	}
}
