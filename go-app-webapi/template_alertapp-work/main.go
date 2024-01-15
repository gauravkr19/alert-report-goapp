// https://stackoverflow.com/questions/52216908/idiomatic-way-to-represent-optional-time-time-in-a-struct
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

	// Render the data using the template and write to the http.ResponseWriter
	err = renderTemplate(w, "template.html", bks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the results as JSON
	// jsonData, err := json.Marshal(bks)
	// if err != nil {
	// 	http.Error(w, "Failed to marshal JSON data", http.StatusInternalServerError)
	// 	return
	// }

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// w.Write(jsonData)
}

func retrieveData(db *sql.DB) ([]Book, error) {
	query := `
	SELECT a.id, a.startsat, a.endsat, a.status, ct.alertname, ct.namespace, ct.priority, ct.severity, ct.deployment, ct.pod, ct.container, ct.replicaset FROM alert a 
	LEFT JOIN (
	SELECT * FROM 
	CROSSTAB('select ct.alertid, ct.label, ct.value FROM AlertLabel ct 
	ORDER BY ct.alertid',
	'VALUES (''alertname''), (''namespace''), (''priority''), (''severity''), (''deployment''), (''pod''), (''container''), (''replicaset'')')
	AS ct (alertid int, alertname VARCHAR, namespace VARCHAR, priority VARCHAR, severity VARCHAR, deployment VARCHAR, pod VARCHAR, container VARCHAR, replicaset VARCHAR) 
	) AS ct ON ct.alertid = a.id;
	`

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
		// if _, ok := bks[entry]; !ok {
		// 	bks[entry] = make(map[string]string)
		// }

		// if nspace.Valid {
		// 	bks[entry]["namespace"] = nspace.String
		// }

	}
	return bks, nil
}

func renderTemplate(w http.ResponseWriter, templateFile string, data interface{}) error {

	tmpl := template.Must(template.ParseFiles(templateFile))

	err := tmpl.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}
