package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/page", pageHandler)
	http.ListenAndServe(":8080", nil)
}
