package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Alertgroup struct {
	Id          int       `json:"id"`
	Time        time.Time `json:time`
	Receiver    string    `json:receiver`
	Status      string    `json:status`
	Externalurl string    `json:externalurl`
	Groupkey    string    `json:groupkey`
}

var db *sql.DB

func init() {
	connStr := "postgres://alertsnitch:alertsnitch@localhost/alertsnitch?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	log.Println("Successfully connected to the database!")
}

func main() {
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/alerts", alertDisplay).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func alertDisplay(w http.ResponseWriter, r *http.Request) {
	var bks []Alertgroup
	rows, err := db.Query("SELECT * FROM alertgroup")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var bk Alertgroup
		err := rows.Scan(&bk.Id, &bk.Time, &bk.Receiver, &bk.Status, &bk.Externalurl, &bk.Groupkey) // order matters
		if err != nil {
			log.Fatal(err)
		}
		bks = append(bks, bk)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bks)
}
