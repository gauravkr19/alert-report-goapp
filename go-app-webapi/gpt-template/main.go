package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://alertsnitch:alertsnitch@localhost/alertsnitch?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")
}

type Book struct {
	Id          int
	Time        time.Time
	Receiver    string
	Status      string
	Externalurl string
	Groupkey    string
}

func main() {
	http.HandleFunc("/books", booksIndex)
	http.ListenAndServe(":8080", nil)
}

func booksIndex(w http.ResponseWriter, r *http.Request) {
	bks, err := retrieveData(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = renderTemplate(w, "template.html", bks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func renderTemplate(w http.ResponseWriter, templateFile string, data interface{}) error {
	tmpl := template.Must(template.ParseFiles(templateFile))
	err := tmpl.Execute(w, data)
	if err != nil {
		return err
	}
	return nil
}

func retrieveData(db *sql.DB) ([]Book, error) {
	rows, err := db.Query("SELECT * FROM alertgroup")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bks := make([]Book, 0)
	for rows.Next() {
		bk := Book{}
		err := rows.Scan(&bk.Id, &bk.Time, &bk.Receiver, &bk.Status, &bk.Externalurl, &bk.Groupkey)
		if err != nil {
			return nil, err
		}
		bks = append(bks, bk)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return bks, nil
}
