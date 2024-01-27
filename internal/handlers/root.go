package handlers

import (
	"errors"
	"log"
	"net/http"
	"text/template"

	"github.com/m4tthewde/truffle/internal/config"
	"github.com/m4tthewde/truffle/internal/session"
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

	_, ok, err := session.SessionFromRequest(r)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			loggedIn = false
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	} else {
		loggedIn = ok
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
