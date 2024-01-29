package main

import (
	"log"
	"net/http"

	"github.com/m4tthewde/truffle/internal/config"
	"github.com/m4tthewde/truffle/internal/handlers"
	"github.com/m4tthewde/truffle/internal/session"
	"github.com/m4tthewde/truffle/internal/twitch"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	session.Init()
	go session.CleanupTicker()

	twitch.InitEventChans(config.Conf.ClientId, config.Conf.ClientSecret)

	handlers.EventChans = make(map[string][]chan twitch.Event)

	rootHandler, err := handlers.NewRootHandler()
	if err != nil {
		log.Fatalln(err)
	}

	chatHandler, err := handlers.NewChatHandler()
	if err != nil {
		log.Fatalln(err)
	}

	chatBoxHandler, err := handlers.NewChatBoxHandler()
	if err != nil {
		log.Fatalln(err)
	}

	settingsHandler, err := handlers.NewSettingsHandler()
	if err != nil {
		log.Fatalln(err)
	}

	wsChatHandler, err := handlers.NewWsChatHandler()
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", rootHandler)
	http.Handle("/chat", chatHandler)
	http.Handle("/chatbox", chatBoxHandler)
	http.Handle("/chat/messages", wsChatHandler)
	http.Handle("/settings", settingsHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/callback/twitch", twitch.CallbackHandler)

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}
