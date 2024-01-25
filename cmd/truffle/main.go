package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/m4tthewde/truffle/internal"
)

func main() {
	godotenv.Load()

	internal.Sessions = make(map[uuid.UUID]int)

	rootHandler, err := internal.NewRootHandler()
	if err != nil {
		log.Fatalln(err)
	}

	dashboardHandler, err := internal.NewDashboardHandler()
	if err != nil {
		log.Fatalln(err)
	}

	settingsHandler, err := internal.NewSettingsHandler()
	if err != nil {
		log.Fatalln(err)
	}

	loginHandler, err := internal.NewLoginHandler()
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", rootHandler)
	http.Handle("/dashboard", dashboardHandler)
	http.Handle("/settings", settingsHandler)
	http.Handle("/login", loginHandler)

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}
