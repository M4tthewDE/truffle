package internal

import (
	"net/http"
	"text/template"
)

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

	sessionId, err := sessionIdFromRequest(r)
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

	err = handler.dashboardTemplate.Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
