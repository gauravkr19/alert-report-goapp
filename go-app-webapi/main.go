package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
)

var db *sql.DB

func init() {
	var err error // check this
	db, err := sql.Open("postgres", "postgres://alertsnitch:alertsnitch@postgresql.monitoring.svc.cluster.local/alertsnitch?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")
}

type AlertGroup struct {
	Id          string
	Time        string
	Receiver    string
	Status      string
	Externalurl string
	Groupkey    string
}

func main() {
	http.HandleFunc("/alerts", alertDisp)
	http.ListenAndServe(":8080", nil)
}

func alertDisp(w http.ResponseWriter, r *http.Request) {
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

	bks := make([]AlertGroup, 0)
	for rows.Next() {
		bk := AlertGroup{}
		err := rows.Scan(&bk.Id, &bk.Time, &bk.Receiver, &bk.Status, &bk.Externalurl, &bk.Groupkey) // order matters
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
		// fmt.Println(bk.Time, bk.Receiver, bk.Status, bk.Externalurl)
		fmt.Fprintf(w, "%s, %s, %s, %s, %s, %s\n", bk.Id, bk.Time, bk.Receiver, bk.Status, bk.Externalurl, bk.Groupkey)
	}

}
