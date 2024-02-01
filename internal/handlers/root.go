package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/m4tthewde/truffle/internal/components"
	"github.com/m4tthewde/truffle/internal/config"
	"github.com/m4tthewde/truffle/internal/session"
)

const authURITemplate = "https://id.twitch.tv/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s/login&scope=user:read:chat channel:moderate"

func RootHandler(w http.ResponseWriter, r *http.Request) {
	_, loggedIn, err := session.SessionFromRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authURI := fmt.Sprintf(authURITemplate, config.Conf.ClientID, config.Conf.URL)
	component := components.Root(loggedIn, templ.URL(authURI))

	err = component.Render(r.Context(), w)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
