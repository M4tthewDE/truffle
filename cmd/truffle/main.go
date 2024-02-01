package main

import (
	"log"
	"net/http"

	"github.com/m4tthewde/truffle/internal/config"
	"github.com/m4tthewde/truffle/internal/handlers"
	"github.com/m4tthewde/truffle/internal/session"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	session.Init()
	go session.CleanupTicker()

	chatHandler, err := handlers.NewChatHandler()
	if err != nil {
		log.Fatalln(err)
	}

	chatRoomHandler, err := handlers.NewChatRoomHandler()
	if err != nil {
		log.Fatalln(err)
	}

	settingsHandler, err := handlers.NewSettingsHandler()
	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/", handlers.RootHandler)
	http.Handle("/chat", chatHandler)
	http.Handle("/chatroom", chatRoomHandler)
	http.HandleFunc("/chat/messages", handlers.WsChatHandler)
	http.Handle("/settings", settingsHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}
