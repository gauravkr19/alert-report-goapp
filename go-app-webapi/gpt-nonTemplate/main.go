package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type TableRow struct {
	Id          int
	Time        time.Time
	Receiver    string
	Status      string
	Externalurl string
	Groupkey    string
}

func main() {
	connStr := "postgres://alertsnitch:alertsnitch@localhost/alertsnitch?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/table", func(w http.ResponseWriter, r *http.Request) {
		data, err := retrieveData(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "<html><body>")
		fmt.Fprintln(w, "<table>")
		fmt.Fprintln(w, "<thead>")
		fmt.Fprintln(w, "<tr>")
		fmt.Fprintln(w, "<th>Date</th>")
		fmt.Fprintln(w, "<th>Integer</th>")
		fmt.Fprintln(w, "<th>String</th>")
		fmt.Fprintln(w, "</tr>")
		fmt.Fprintln(w, "</thead>")
		fmt.Fprintln(w, "<tbody>")
		for _, row := range data {
			fmt.Fprintf(w, "<tr>")
			fmt.Fprintf(w, "<td>%d</td>", row.Id)
			fmt.Fprintf(w, "<td>%s</td>", row.Time)
			fmt.Fprintf(w, "<td>%s</td>", row.Receiver)
			fmt.Fprintf(w, "<td>%s</td>", row.Status)
			fmt.Fprintf(w, "<td>%s</td>", row.Externalurl)
			fmt.Fprintf(w, "<td>%s</td>", row.Groupkey)
			fmt.Fprintf(w, "</tr>")
		}
		fmt.Fprintln(w, "</tbody>")
		fmt.Fprintln(w, "</table>")
		fmt.Fprintln(w, "</body></html>")
	})

	http.ListenAndServe(":8080", nil)
}

func retrieveData(db *sql.DB) ([]TableRow, error) {
	query := "SELECT * FROM alertgroup"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []TableRow
	for rows.Next() {
		var rowData TableRow
		err := rows.Scan(&rowData.Id, &rowData.Time, &rowData.Receiver, &rowData.Status, &rowData.Externalurl, &rowData.Groupkey)
		if err != nil {
			return nil, err
		}
		data = append(data, rowData)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}
