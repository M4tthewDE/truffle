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

	http.HandleFunc("/", handlers.RootHandler)
	http.HandleFunc("/chat", handlers.ChatHandler)
	http.HandleFunc("/chatroom", handlers.ChatRoomHandler)
	http.HandleFunc("/chat/messages", handlers.WsChatHandler)
	http.HandleFunc("/settings", handlers.SettingsHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}
