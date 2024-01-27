package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/m4tthewde/truffle/internal"
	"github.com/m4tthewde/truffle/internal/twitch"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := internal.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	internal.Sessions = make(map[uuid.UUID]internal.UserInfo)
	internal.EventChans = make(map[string][]chan twitch.Event)

	rootHandler, err := internal.NewRootHandler()
	if err != nil {
		log.Fatalln(err)
	}

	chatHandler, err := internal.NewChatHandler()
	if err != nil {
		log.Fatalln(err)
	}

	settingsHandler, err := internal.NewSettingsHandler()
	if err != nil {
		log.Fatalln(err)
	}

	wsChatHandler, err := internal.NewWsChatHandler()
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", rootHandler)
	http.Handle("/chat", chatHandler)
	http.Handle("/chat/messages", wsChatHandler)
	http.Handle("/settings", settingsHandler)
	http.HandleFunc("/login", internal.LoginHandler)
	http.HandleFunc("/logout", internal.LogoutHandler)

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}
