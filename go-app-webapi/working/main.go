// https://stackoverflow.com/questions/52216908/idiomatic-way-to-represent-optional-time-time-in-a-struct
package main

import (
	"alertapp-working/pkg/database"
	"alertapp-working/pkg/handlers"
	"net/http"
)

func main() {
	// Calling DB to ensure the init function is executed
	_ = database.DB

	// Register the BooksIndex handler with the DB instance
	http.HandleFunc("/books", handlers.BooksIndex(database.DB))
	http.ListenAndServe(":8080", nil)
}
