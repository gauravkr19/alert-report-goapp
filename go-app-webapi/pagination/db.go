package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	_ "github.com/lib/pq"
)

// Define pagination parameters
const (
	pageSize    = 50 // Number of records per page
	defaultPage = 1  // Default page number
)

// YourDataStruct represents your data model. Update this as needed.
type YourDataStruct struct {
	ID         int
	StartsAt   string
	EndsAt     string
	Status     string
	AlertName  string
	Namespace  string
	Priority   string
	Severity   string
	Deployment string
	Pod        string
	Container  string
	Replicaset string
}

// PageData represents data and pagination information.
type PageData struct {
	Data        []YourDataStruct
	CurrentPage int
	TotalPages  int
}

// Initialize the database connection.
func initDB() (*sql.DB, error) {
	connectionString := "user=myuser dbname=mydb sslmode=disable"
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Fetch data for the specified page.
func fetchDataForPage(db *sql.DB, page, pageSize int) ([]YourDataStruct, error) {
	// Calculate LIMIT and OFFSET based on pagination parameters
	limit := pageSize
	offset := (page - 1) * pageSize

	// Update your SQL query with LIMIT and OFFSET
	query := fmt.Sprintf(`
        SELECT a.id, a.startsat, a.endsat, a.status, ct.alertname, ct.namespace, ct.priority, ct.severity, ct.deployment, ct.pod, ct.container, ct.replicaset FROM alert a
        LEFT JOIN (
            SELECT * FROM 
            CROSSTAB('select ct.alertid, ct.label, ct.value FROM AlertLabel ct  
            ORDER BY ct.alertid',
            'VALUES (''alertname''), (''namespace''), (''priority''), (''severity''), (''deployment''), (''pod''), (''container''), (''replicaset'')')
            AS ct (alertid int, alertname VARCHAR, namespace VARCHAR, priority VARCHAR, severity VARCHAR, deployment VARCHAR, pod VARCHAR, container VARCHAR, replicaset VARCHAR) 
        ) AS ct ON ct.alertid = a.id
        LIMIT %d OFFSET %d;
    `, limit, offset)

	// Execute the query and retrieve data
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []YourDataStruct
	for rows.Next() {
		var d YourDataStruct
		err := rows.Scan(
			&d.ID, &d.StartsAt, &d.EndsAt, &d.Status, &d.AlertName,
			&d.Namespace, &d.Priority, &d.Severity, &d.Deployment,
			&d.Pod, &d.Container, &d.Replicaset,
		)
		if err != nil {
			return nil, err
		}
		data = append(data, d)
	}

	// Calculate total records (update as needed)
	totalRecords := 12000
	// Calculate total pages
	totalPages := (totalRecords + pageSize - 1) / pageSize

	return data, nil
}

// Handler for the /page route
func pageHandler(w http.ResponseWriter, r *http.Request) {
	// Initialize the database connection
	db, err := initDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Parse page number from query parameter
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = defaultPage
	}

	// Fetch data for the specified page
	data, err := fetchDataForPage(db, page, pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a PageData struct with data and pagination information
	pageData := PageData{
		Data:        data,
		CurrentPage: page,
		TotalPages:  totalPages,
	}

	// Parse and execute the HTML template
	tmpl, err := template.ParseFiles("templates/page.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, pageData)
}
