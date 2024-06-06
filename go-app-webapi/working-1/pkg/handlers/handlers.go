package handlers

import (
	"alertapp-working/pkg/database"
	"alertapp-working/pkg/models"
	"alertapp-working/pkg/render"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Define WebSocket message struct
type WebSocketMessage struct {
	MessageType string `json:"messageType"`
	// Data        []models.Book `json:"data"`
	Data      interface{} `json:"data"`
	DateRange string
}

// Clients is a map to keep track of connected clients
var clients = make(map[*websocket.Conn]bool)

// Broadcast channel to send messages to the clients
var broadcast = make(chan WebSocketMessage)

// upgradeConnection is the websocket upgrader from gorilla/websockets
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Store the date range in a global variable or a session (for simplicity, using a global variable here)
var globalDateRange string

func BooksIndex(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// func BooksIndex(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters, fmt.Sscanf to parse the value of the "page" parameter as an integer and assigns it to the page variable.
		page := 1
		if r.URL.Query().Get("page") != "" {
			fmt.Sscanf(r.URL.Query().Get("page"), "%d", &page)
		}

		// Calculate limit and offset based on page number
		limit := 30
		offset := (page - 1) * limit

		// Count total number of records in the table
		totalRecords, err := database.CountTotalRecords(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Calculate total number of pages
		totalPages := (totalRecords + limit - 1) / limit

		// Check if offset is out of range
		if offset < 0 || offset >= totalRecords {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}

		// Fetch data for the specified page
		bks, err := database.FetchData(db, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// This code iterates over each key-value pair in r.URL.Query(), extracts the first value from the slice of values, and stores it in the queryParams map.
		// Then, queryParams is assigned to the QueryParams field of the PageData struct.
		queryParams := make(map[string]string)
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = values[0]
			}
		}

		// Prepare the data.
		pageData := models.PageData{
			Books:       bks,
			Page:        page,
			TotalPages:  totalPages,
			QueryParams: queryParams,
		}

		// Render the data using the template and write to the http.ResponseWriter
		err = render.RenderTemplate(w, "books_template.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// isToday - helper func  to determine if the selected date range is "Today"
// func isToday(startTime, endTime time.Time) bool {
// 	now := time.Now()
// 	return startTime.Before(now) && endTime.After(now)
// }

// parses the daterange from the daterange user input form
func parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	dateRange := r.FormValue("daterange")
	globalDateRange = dateRange
	fmt.Println("parseDateRange-Date Range from Form:", dateRange)

	dates := strings.Split(dateRange, " - ")
	if len(dates) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date range format")
	}

	fmt.Println("parseDateRange:", dates[0])

	// TimeZone is added with time.Time, so remove it and convert to string.

	startDate, err := time.Parse("2006-01-02 15:04:05", dates[0])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endDate, err := time.Parse("2006-01-02 15:04:05", dates[1])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startDate, endDate, nil
}

// parseDateRangeFromString parses date range from string
func parseDateRangeFromString(dateRange string) (time.Time, time.Time, error) {
	dates := strings.Split(dateRange, " - ")
	if len(dates) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date range format")
	}

	startTime, err := time.Parse("2006-01-02 15:04:05", dates[0])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endTime, err := time.Parse("2006-01-02 15:04:05", dates[1])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startTime, endTime, nil
}

// func Today(startTime, endTime time.Time) bool {
// 	now := time.Now()
// 	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// 	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Second)

// 	return startTime.After(startOfDay) && endTime.Before(endOfDay)
// }

// ExportBooks is an HTTP handler function to export books data to Excel
func ExportBooks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		startTime, endTime, err := parseDateRange(r)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing date range: %v", err), http.StatusBadRequest)
			return
		}

		// Check if the selected range is "Today"
		isToday := startTime.Format("2006-01-02") == time.Now().Format("2006-01-02")

		// isToday := startTime.Equal(time.Now().Truncate(24*time.Hour)) && endTime.Equal(time.Now().Truncate(24*time.Hour).Add(24*time.Hour-1))
		if !isToday {
			// If the selected range is not "Today", proceed with regular HTTP request handling
			bks, err := database.FetchBooksByTimeRange(db, startTime, endTime)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			exportData := models.ExportData{
				Books: bks,
			}

			// Render the data using the template to export in XLS
			err = render.RenderTemplate(w, "excel_template.html", exportData)
			if err != nil {
				log.Println("Error rendering template:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		// If the selected range is "Today", upgrade the connection to websockets and handle real-time updates
		if isToday {
			fmt.Println("isToday block")
			// If the selected date range is today, call the WebSocket handler
			http.Redirect(w, r, "/ws", http.StatusSeeOther)
			// If the selected date range is today, redirect to the WebSocket handler with query parameters
			// http.Redirect(w, r, fmt.Sprintf("/ws?start=%s&end=%s", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339)), http.StatusSeeOther)
			return
		}
		// request with payload - hand that off to another goroutine that listens to a channel
		// and does different things depending of payload content.
	}
}

// WebSocketPostHandler for handling the POST request to parse the date range and redirect
func WebSocketPostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		dateRange := r.FormValue("daterange")
		fmt.Println("WebSocketPostHandler - Recd date range:", dateRange)
		globalDateRange = dateRange

		// Parse the date range from the form data
		// startTime, endTime, err := parseDateRange(r)
		// if err != nil {
		// 	http.Error(w, "Invalid date range", http.StatusBadRequest)
		// 	return
		// }

		fmt.Println("WebSocketPostHandler", globalDateRange)

		// Redirect to the WebSocket GET handler with query parameters
		// http.Redirect(w, r, fmt.Sprintf("/ws?start=%s&end=%s", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339)), http.StatusSeeOther)
	}
}

// Function to handle WebSocket connections
func WebSocketHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Upgrade the connection to a WebSocket
		conn, err := upgradeConnection.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error upgrading to WebSocket:", err)
			return
		}

		defer func() {
			conn.Close()
			delete(clients, conn)
		}()

		// Register new client
		clients[conn] = true
		var startTime, endTime time.Time

		// Start SendDataToClients function as a Goroutine
		// go SendDataToClients(db, startTime, endTime)

		// Infinite loop to continuously send data to client
		go func() {
			for {
				msg := <-broadcast
				for client := range clients {
					err := client.WriteJSON(msg)
					if err != nil {
						log.Printf("Error writing to WebSocket client: %v", err)
						client.Close()
						delete(clients, client)
					}
				}
			}
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading WebSocket message: %v", err)
				delete(clients, conn)
				break
			}

			var msg WebSocketMessage
			err = json.Unmarshal(message, &msg)
			if err != nil {
				log.Printf("Error unmarshalling WebSocket message: %v", err)
				continue
			}

			if msg.MessageType == "init" {
				startTime, endTime, err = parseDateRangeFromString(msg.DateRange)
				if err != nil {
					log.Printf("Error parsing date range: %v", err)
					continue
				}

				go SendDataToClients(db, startTime, endTime)
			}
		}

	}
}

// Function to send data to WebSocket clients
func SendDataToClients(db *sql.DB, startTime, endTime time.Time) {

	fmt.Printf("SendDataToClients:%v \n", startTime)

	// Test data: array of strings
	// records := []string{"Record 1", "Record 2", "Record 3"}

	// Infinite loop to continuously send data to clients
	for {
		records, err := database.FetchBooksByTimeRange(db, startTime, endTime)
		if err != nil {
			log.Println("Error fetching records from database:", err)
			continue
		}

		// Create WebSocket message
		message := WebSocketMessage{
			MessageType: "update",
			Data:        records,
		}

		// Send message to all connected clients
		broadcast <- message

		// Sleep for some time before fetching records again
		time.Sleep(5 * time.Second)
	}
}

func Home(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The if block (POST Request) is NOT executed, as we are calling /export from form action in template after taking the user input
		if r.Method == http.MethodPost {
			// Extract the value of the daterange field from the form submission
			var err error
			dateRange := r.FormValue("daterange")
			if err != nil {
				http.Error(w, "Error parsing form", http.StatusBadRequest)
				return
			}

			// Split the daterange value to get the start and end dates
			dates := strings.Split(dateRange, " - ")
			if len(dates) != 2 {
				http.Error(w, "Invalid date range", http.StatusBadRequest)
				return
			}
			startDate := dates[0]
			endDate := dates[1]

			// Do something with the start and end dates (e.g., save to database, perform calculations)
			fmt.Println("Start Date:", startDate)
			fmt.Println("End Date:", endDate)

			// Return the dates as JSON response
			data := struct {
				StartDate string `json:"startDate"`
				EndDate   string `json:"endDate"`
			}{
				StartDate: startDate,
				EndDate:   endDate,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
			return
		}

		// Render the template to get user input with GET request (non POST request)
		err := render.RenderTemplate(w, "home_template.html", nil)
		if err != nil {
			fmt.Println("Error with rendering")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
