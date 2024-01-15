package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	var err error
	db, err := sql.Open("postgres", "postgres://alertsnitch:alertsnitch@localhost/alertsnitch?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	fmt.Println("You connected to your database.")

	query := `
	SELECT a.id, ct.alertname, ct.namespace FROM alert a 
	LEFT JOIN (
	SELECT * FROM 
	CROSSTAB('select ct.alertid, ct.label, ct.value FROM AlertLabel ct 
	ORDER BY ct.alertid',
	'VALUES (''alertname''), (''namespace'')')
	AS ct (alertid int, alertname VARCHAR, namespace VARCHAR) 
	) AS ct ON ct.alertid = a.id;
	`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	books := make(map[int]map[string]string)

	for rows.Next() {
		var entry int
		var alname, nspace sql.NullString
		err := rows.Scan(&entry, &alname, &nspace)
		if err != nil {
			log.Fatal(err)
		}

		// Store the data in the resultMap
		if _, ok := books[entry]; !ok {
			books[entry] = make(map[string]string)
		}

		if alname.Valid {
			books[entry]["alertname"] = alname.String
		}
		if nspace.Valid {
			books[entry]["namespace"] = nspace.String
		}

		// bks = append(bks, entry)
	}

	for entry, labels := range books {
		fmt.Printf("Entry: %d. Alertname: %s, namespaceL %s \n", entry, labels["alertname"], labels["namespace"])
		// fmt.Fprintf(w, "%d, %s, %s, %s, %s, %s\n", &entry.Id, &entry.Startsat, &entry.Status, &entry.Labels)
	}
}
