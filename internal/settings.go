package internal

import (
	"net/http"
	"text/template"
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

	err := handler.settingsTemplate.Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
