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
	DateRange string      `json:"dateRange"`
}

// Clients is a map to keep track of connected clients
var clients = make(map[*websocket.Conn]bool)

// Broadcast channel to send messages to the clients
var broadcast = make(chan WebSocketMessage)

// Channel to receive WebSocket messages
var wsChan = make(chan WebSocketMessage)

// Define upgradeConnection to upgrade the HTTP connection to WebSocket
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

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

// parses the daterange from the daterange user input form
func parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	dateRange := r.FormValue("daterange")

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

	startDate, err := time.Parse("2006-01-02 15:04:05", dates[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format: %v", err)
	}

	endDate, err := time.Parse("2006-01-02 15:04:05", dates[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format: %v", err)
	}

	return startDate, endDate, nil
}

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
	}
}

// Upgrade connection and call goroutine ListenForWs
func WebSocketHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradeConnection.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error upgrading to WebSocket:", err)
			return
		}
		defer conn.Close()

		// Register new client
		clients[conn] = true
		log.Println("WebSocket connection established.")

		// Start goroutine to listen for messages from WebSocket, and send to channel
		go ListenForWs(conn)

		// Listen for messages from the broadcast channel
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

	}
}

// Function to listen for WebSocket messages
func ListenForWs(conn *websocket.Conn) {
	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading from WebSocket:", err)
			delete(clients, conn)
			break
		}
		// Send the received message to the channel
		wsChan <- msg
	}
}

// Function to listen to wsChan channel
func ListenToWsChannel(db *sql.DB) {
	for {
		msg := <-wsChan
		if msg.MessageType == "dateRange" {
			startTime, endTime, err := parseDateRangeFromString(msg.DateRange)
			if err != nil {
				log.Println("Error parsing date range:", err)
				continue
			}
			log.Printf("Received date range - Start: %v, End: %v\n", startTime, endTime)

			// Fetch data and broadcast to all clients
			data := fetchData(db, startTime, endTime)
			broadcastToAll(WebSocketMessage{
				MessageType: "update",
				Data:        data,
				DateRange:   msg.DateRange, // Implement fetchData to get data based on date range
			})
		}
	}
}

// Function to broadcast messages to all connected clients
func broadcastToAll(msg WebSocketMessage) {
	broadcast <- msg
}

// Implement fetchData to get data based on date range
func fetchData(db *sql.DB, startTime, endTime time.Time) interface{} {
	// Fetch data from the database based on the start and end times
	// Return the fetched data
	books, err := database.FetchBooksByTimeRange(db, startTime, endTime)
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		return nil
	}

	if len(books) == 0 {
		log.Println("No records found for the given date range")
		return "NoData"
	}

	log.Printf("Fetched %d books from database\n", len(books))
	return books
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
