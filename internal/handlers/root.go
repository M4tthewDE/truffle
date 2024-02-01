package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/m4tthewde/truffle/internal/components"
	"github.com/m4tthewde/truffle/internal/config"
	"github.com/m4tthewde/truffle/internal/session"
)

const authUriTemplate = "https://id.twitch.tv/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s/login&scope=user:read:chat channel:moderate"

func RootHandler(w http.ResponseWriter, r *http.Request) {
	_, loggedIn, err := session.SessionFromRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authUri := fmt.Sprintf(authUriTemplate, config.Conf.ClientId, config.Conf.Url)
	component := components.Root(loggedIn, templ.URL(authUri))
	err = component.Render(context.Background(), w)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
