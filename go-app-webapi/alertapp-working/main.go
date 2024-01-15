// https://stackoverflow.com/questions/52216908/idiomatic-way-to-represent-optional-time-time-in-a-struct
package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// Define pagination parameters
const (
	pageSize    = 50 // Number of records per page
	defaultPage = 1  // Default page number
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://alertsnitch:alertsnitch@localhost/alertsnitch?sslmode=disable")
	// db, err = sql.Open("postgres", "postgres://alertsnitch:alertsnitch@35.205.240.171/alertsnitch?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")
}

type Book struct {
	Id         int
	Startsat   *time.Time
	Endsat     *time.Time
	Status     string
	Alertname  string
	Namespace  string
	Priority   string
	Severity   string
	Deployment string
	Pod        string
	Container  string
	Replicaset string
}

type PageData struct {
	Data        []Book
	CurrentPage int
	TotalPages  int
	PrevPage    int
	NextPage    int
	PageNumbers []int
}

func main() {
	http.HandleFunc("/books", booksIndex)
	http.ListenAndServe(":8080", nil)
}

func booksIndex(w http.ResponseWriter, r *http.Request) {

	// Parse page number from query parameter
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = defaultPage
	}

	// Calculate total records (update as needed)
	totalRecords := 12000
	// Calculate total pages
	totalPages := (totalRecords + pageSize - 1) / pageSize

	// Fetch data for the specified page
	bks, err := retrieveData(db, page, pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate previous and next page numbers
	var pageData PageData
	prevPage := pageData.CurrentPage - 1
	nextPage := pageData.CurrentPage + 1

	// Create a PageData struct with data and pagination information
	pageData = PageData{
		Data:        bks,
		CurrentPage: page,
		TotalPages:  totalPages,
		PrevPage:    prevPage,
		NextPage:    nextPage,
		PageNumbers: generatePageNumbers(page, totalPages),
	}

	// Render the data using the template and write to the http.ResponseWriter
	if _, err := renderTemplate(w, pageData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func retrieveData(db *sql.DB, page, pageSize int) ([]Book, error) {

	// Calculate LIMIT and OFFSET based on pagination parameters
	limit := pageSize
	offset := (page - 1) * pageSize

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

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bks []Book

	for rows.Next() {
		var id int
		var alname, nspace, prio, sever, deplo, pod, conta, repli, stat sql.NullString
		var end, start sql.NullTime
		err := rows.Scan(&id, &start, &end, &stat, &alname, &nspace, &prio, &sever, &deplo, &pod, &conta, &repli)
		if err != nil {
			return nil, err
		}

		var endsat, startsat *time.Time
		if end.Valid {
			endsat = &end.Time
		}
		if start.Valid {
			startsat = &start.Time
		}

		bks = append(bks, Book{
			Id:         id,
			Startsat:   startsat,
			Endsat:     endsat,
			Status:     stat.String,
			Alertname:  alname.String,
			Namespace:  nspace.String,
			Priority:   prio.String,
			Severity:   sever.String,
			Deployment: deplo.String,
			Pod:        pod.String,
			Container:  conta.String,
			Replicaset: repli.String,
		})

	}
	return bks, nil
}

func renderTemplate(w http.ResponseWriter, pageData PageData) (string, error) {

	// tmpl := template.Must(template.ParseFiles(templateFile))

	// err := tmpl.Execute(w, pageData)
	// if err != nil {
	// 	return err
	// }

	// return nil

	tmpl, err := template.New("page").Funcs(template.FuncMap{
		"Seq": Seq,
	}).ParseFiles("template.html")
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "template.html", pageData)
	if err != nil {
		return "", err
	}

	return buf.String(), nil

}

func Seq(start, end int) []int {
	var s []int
	for i := start; i <= end; i++ {
		s = append(s, i)
	}
	return s
}

func generatePageNumbers(currentPage, totalPages int) []int {
	var pageNumbers []int

	// Define how many page numbers you want to show in the pagination bar
	maxPageNumbers := 5

	// Calculate the middle point of the page numbers
	middle := maxPageNumbers / 2

	// Calculate the starting and ending page numbers
	start := currentPage - middle
	if start < 1 {
		start = 1
	}

	end := start + maxPageNumbers - 1
	if end > totalPages {
		end = totalPages
		start = end - maxPageNumbers + 1
		if start < 1 {
			start = 1
		}
	}

	// Create a slice of page numbers to display
	for i := start; i <= end; i++ {
		pageNumbers = append(pageNumbers, i)
	}

	return pageNumbers
}
