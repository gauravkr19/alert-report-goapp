package handlers

import (
	"alertapp-working/pkg/database"
	"alertapp-working/pkg/models"
	"alertapp-working/pkg/render"
	"database/sql"
	"fmt"
	"net/http"
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
		limit := 20
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

		// Calculate the end index for the subset slice
		// end := offset + limit
		// if end > len(bks) {
		// 	end = len(bks)
		// }

		// Slice books to display only the subset defined by offset and limit
		// subsetBooks := bks[offset:end]

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
		err = render.RenderTemplate(w, "template.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
