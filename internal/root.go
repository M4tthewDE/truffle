package internal

import (
	"log"
	"net/http"
	"os"
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
	// TODO: get logged in status from db using sessionid from request
	rootData := RootData{
		LoggedIn: false,
		ClientId: os.Getenv("CLIENT_ID"),
	}
	err := handler.indexTemplate.Execute(w, rootData)
	if err != nil {
		log.Fatalln(err)
	}
}
