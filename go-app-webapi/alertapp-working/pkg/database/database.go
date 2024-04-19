package database

import (
	"database/sql"
	"fmt"
	"time"

	"alertapp-working/pkg/models"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func init() {
	var err error
	DB, err = sql.Open("postgres", "postgres://alertsnitch:alertsnitch@localhost/alertsnitch?sslmode=disable")
	// DB, err = sql.Open("postgres", "postgres://alertsnitch:alertsnitch@35.205.240.171/alertsnitch?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = DB.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")
}

// FetchData performs the DB query and assigns the value to Go struct bks
func FetchData(DB *sql.DB, limit, offset int) ([]models.Book, error) {
	// Check if offset is out of range
	if offset < 0 {
		return []models.Book{}, nil
	}

	query := `
    SELECT a.id, a.fingerprint, a.startsat, a.endsat, a.status, ct.alertname, ct.namespace, ct.priority, ct.severity, ct.deployment, ct.pod, ct.container, ct.replicaset FROM alert a
    LEFT JOIN (
        SELECT * FROM 
        CROSSTAB('select ct.alertid, ct.label, ct.value FROM AlertLabel ct  
        ORDER BY ct.alertid',
        'VALUES (''alertname''), (''namespace''), (''priority''), (''severity''), (''deployment''), (''pod''), (''container''), (''replicaset'')')
        AS ct (alertid int, alertname VARCHAR, namespace VARCHAR, priority VARCHAR, severity VARCHAR, deployment VARCHAR, pod VARCHAR, container VARCHAR, replicaset VARCHAR) 
    ) AS ct ON ct.alertid = a.id
    LIMIT $1 OFFSET $2;
	`

	rows, err := DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bks []models.Book

	for rows.Next() {
		var id int
		// var fingerprint string
		var alertname, namespace, priority, severity, deployment, pod, container, replicaset, status, fingerprint sql.NullString
		var end, start sql.NullTime
		err := rows.Scan(&id, &fingerprint, &start, &end, &status, &alertname, &namespace, &priority, &severity, &deployment, &pod, &container, &replicaset)
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

		bks = append(bks, models.Book{
			Id:          id,
			Fingerprint: fingerprint.String,
			Startsat:    startsat,
			Endsat:      endsat,
			Status:      status.String,
			Alertname:   alertname.String,
			Namespace:   namespace.String,
			Priority:    priority.String,
			Severity:    severity.String,
			Deployment:  deployment.String,
			Pod:         pod.String,
			Container:   container.String,
			Replicaset:  replicaset.String,
		})

	}
	return bks, nil
}

// countTotalRecords get the count of total records
func CountTotalRecords(DB *sql.DB) (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM alert").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
