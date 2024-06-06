// https://stackoverflow.com/questions/52216908/idiomatic-way-to-represent-optional-time-time-in-a-struct
package main

import (
	"alertapp-working/pkg/database"
	"alertapp-working/pkg/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Calling DB to ensure the init function is executed
	_ = database.DB

	// Start listening for WebSocket messages
	go handlers.ListenToWsChannel(database.DB)
	// go handlers.BroadcastToAll(handlers.WebSocketMessage)

	r := mux.NewRouter()

	go handlers.ListenToWsChannel(database.DB)

	// Endpoint for rendering the initial page with the date range picker
	r.HandleFunc("/home", handlers.Home(database.DB)).Methods("GET", "POST")

	r.HandleFunc("/books", handlers.BooksIndex(database.DB)).Methods("GET")

	// Endpoint for handling form submission
	r.HandleFunc("/export", handlers.ExportBooks(database.DB)).Methods("POST")

	r.HandleFunc("/ws", handlers.WebSocketHandler(database.DB)).Methods("GET")

	http.ListenAndServe(":8080", r)

}
