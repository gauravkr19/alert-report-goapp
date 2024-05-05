package models

import "time"

type Book struct {
	Id          int
	Fingerprint string
	Startsat    *time.Time
	Endsat      *time.Time
	Status      string
	Alertname   string
	Namespace   string
	Priority    string
	Severity    string
	Deployment  string
	Pod         string
	Container   string
	Replicaset  string
}

type PageData struct {
	Books       []Book
	Page        int
	Limit       int
	TotalPages  int
	QueryParams map[string]string
}

type ExportData struct {
	Books []Book
}
