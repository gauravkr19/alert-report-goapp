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

	r := mux.NewRouter()

	// Endpoint for rendering the initial page with the date range picker
	r.HandleFunc("/home", handlers.Home(database.DB)).Methods("GET")
	r.HandleFunc("/home", handlers.Home(database.DB)).Methods("POST")

	// Endpoint for handling form submission
	r.HandleFunc("/export", handlers.ExportBooks(database.DB)).Methods("POST")

	http.ListenAndServe(":8080", r)

	// Register the BooksIndex handler with the DB instance
	// http.HandleFunc("/books", handlers.BooksIndex(database.DB))
	// http.HandleFunc("/export", handlers.ExportBooks(database.DB))
	// http.HandleFunc("/home", handlers.Home(database.DB))
	// http.ListenAndServe(":8080", nil)
}
