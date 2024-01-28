package handlers

import (
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
	Url      string
}

func (handler *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, loggedIn, err := session.SessionFromRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rootData := RootData{
		LoggedIn: loggedIn,
		ClientId: config.Conf.ClientId,
		Url:      config.Conf.Url,
	}
	err = handler.indexTemplate.Execute(w, rootData)
	if err != nil {
		log.Println(err)
	}
}
