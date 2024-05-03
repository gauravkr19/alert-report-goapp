package handlers

import (
	"alertapp-working/pkg/database"
	"alertapp-working/pkg/models"
	"alertapp-working/pkg/render"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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

// ExportBooks is an HTTP handler function to export books data to Excel
func ExportBooks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// if r.Method != http.MethodPost {
		// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		// 	return
		// }

		// Parse form data to get the start and end dates
		startDate := r.FormValue("startDate")
		endDate := r.FormValue("endDate")

		// Extract start and end dates from request data
		startTime, err := time.Parse("2006-01-02 15:04:05.999", startDate)
		if err != nil {
			http.Error(w, "Invalid start date", http.StatusBadRequest)
			return
		}

		endTime, err := time.Parse("2006-01-02 15:04:05.999", endDate)
		if err != nil {
			http.Error(w, "Invalid end date", http.StatusBadRequest)
			return
		}

		fmt.Println("Handler's starttime is", startTime)

		// Fetch data from the database based on the time range
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
			fmt.Println("Error with rendering")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func Home(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is POST
		if r.Method == http.MethodPost {
			// Parse form data
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Error parsing form", http.StatusBadRequest)
				return
			}

			// Get start and end dates from form data
			startDate := r.Form.Get("startDate")
			endDate := r.Form.Get("endDate")

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

		// Render the template to get user input
		err := render.RenderTemplate(w, "home_template.html", nil)
		if err != nil {
			fmt.Println("Error with rendering")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
