package handlers

import (
	"net/http"
	"text/template"

	"github.com/m4tthewde/truffle/internal/util"
)

type SettingsHandler struct {
	settingsTemplate *template.Template
}

func NewSettingsHandler() (*SettingsHandler, error) {
	settingsTemplate, err := template.ParseFiles("resources/settings.html")
	if err != nil {
		return nil, err
	}

	return &SettingsHandler{settingsTemplate: settingsTemplate}, nil
}

func (handler *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionId, err := util.SessionIdFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		_, loggedIn := Sessions[*sessionId]
		if !loggedIn {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	err = handler.settingsTemplate.Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
