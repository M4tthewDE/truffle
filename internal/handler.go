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
	rootData := RootData{
		LoggedIn: false,
		ClientId: os.Getenv("CLIENT_ID"),
	}
	err := handler.indexTemplate.Execute(w, rootData)
	if err != nil {
		log.Fatalln(err)
	}
}

type DashboardHandler struct {
	dashboardTemplate *template.Template
}

func NewDashboardHandler() (*DashboardHandler, error) {
	dashboardTemplate, err := template.ParseFiles("resources/dashboard.html")
	if err != nil {
		return nil, err
	}

	return &DashboardHandler{dashboardTemplate: dashboardTemplate}, nil
}

func (handler *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := handler.dashboardTemplate.Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

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
