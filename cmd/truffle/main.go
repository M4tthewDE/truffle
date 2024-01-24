package main

import (
	"github.com/joho/godotenv"
	"github.com/m4tthewde/truffle/internal"
	"log"
	"net/http"
)

func main() {
	godotenv.Load()

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
