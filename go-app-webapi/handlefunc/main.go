package main

import (
	"database/sql"
	"fmt"
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
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM alertgroup")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	bks := make([]Book, 0)
	for rows.Next() {
		bk := Book{}
		err := rows.Scan(&bk.Id, &bk.Time, &bk.Receiver, &bk.Status, &bk.Externalurl, &bk.Groupkey)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		bks = append(bks, bk)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for _, bk := range bks {
		fmt.Fprintf(w, "%d, %s, %s, %s, %s, %s\n", bk.Id, bk.Time, bk.Receiver, bk.Status, bk.Externalurl, bk.Groupkey)
	}
}
