package main

import (
	"github.com/m4tthewde/truffle/internal"
	"log"
	"net/http"
)

func main() {
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

	http.Handle("/", rootHandler)
	http.Handle("/dashboard", dashboardHandler)
	http.Handle("/settings", settingsHandler)

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}
