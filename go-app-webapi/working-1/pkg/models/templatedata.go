package models

import "time"

type Book struct {
	Id          int        `json:"id"`
	Fingerprint string     `json:"fingerprint"`
	Startsat    *time.Time `json:"startsat"`
	Endsat      *time.Time `json:"endsat"`
	Status      string     `json:"status"`
	Alertname   string     `json:"alertname"`
	Namespace   string     `json:"namespace"`
	Priority    string     `json:"priority"`
	Severity    string     `json:"severity"`
	Deployment  string     `json:"deployment"`
	Pod         string     `json:"pod"`
	Container   string     `json:"container"`
	Replicaset  string     `json:"replicaset"`
}

type ExportData struct {
	Books []Book `json:"books"`
}

type PageData struct {
	Books       []Book
	Page        int
	Limit       int
	TotalPages  int
	QueryParams map[string]string
}
