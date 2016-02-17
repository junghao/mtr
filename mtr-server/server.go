package main

import (
	"database/sql"
	_ "github.com/GeoNet/cfg/cfgenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"time"
)

var mux *http.ServeMux
var db *sql.DB
var dbR *sql.DB // Database connection with read only credentials
var userW, keyW string
var userR, keyR string

var eol = []byte("\n")

func init() {
	userW = os.Getenv("MTR_USER")
	keyW = os.Getenv("MTR_KEY")
	userR = os.Getenv("MTR_USER_R")
	keyR = os.Getenv("MTR_KEY_R")

	mux = http.NewServeMux()
	mux.HandleFunc("/field/locality", auth(localityHandler))
	mux.HandleFunc("/field/locality/dark", auth(localityDarkHandler))
	mux.HandleFunc("/field/source", auth(sourceHandler))
	mux.HandleFunc("/field/metric", auth(fieldMetricHandler))
	mux.HandleFunc("/field/metric/summary", auth(fieldMetricSummaryHandler))
	mux.HandleFunc("/field/metric/threshold", auth(thresholdHandler))
	mux.HandleFunc("/field/metric/tag", auth(tagHandler))
	mux.HandleFunc("/field/metric/type", auth(typeHandler))
	mux.HandleFunc("/health", health)
}

func main() {
	var err error
	db, err = sql.Open("postgres",
		os.ExpandEnv("host=${DB_HOST} connect_timeout=30 user=${DB_USER} password=${DB_PASSWORD} dbname=mtr sslmode=disable"))
	if err != nil {
		log.Println("Problem with DB config.")
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)

	if err = db.Ping(); err != nil {
		log.Println("ERROR: problem pinging DB - is it up and contactable? 500s will be served")
	}

	dbR, err = sql.Open("postgres",
		os.ExpandEnv("host=${DB_HOST} connect_timeout=30 user=${DB_USER_R} password=${DB_PASSWORD_R} dbname=mtr sslmode=disable"))
	if err != nil {
		log.Println("Problem with DB config.")
		log.Fatal(err)
	}
	defer dbR.Close()

	dbR.SetMaxIdleConns(20)
	dbR.SetMaxOpenConns(20)

	if err = dbR.Ping(); err != nil {
		log.Println("ERROR: problem pinging DB - is it up and contactable? 500s will be served")
	}

	go refreshMetrics()

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func auth(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT", "DELETE":
			if user, password, ok := r.BasicAuth(); ok && userW == user && keyW == password {
				f(w, r)
			} else {
				http.Error(w, "Access denied", http.StatusUnauthorized)
				return
			}
		case "GET":
			if user, password, ok := r.BasicAuth(); ok && userR == user && keyR == password {
				f(w, r)
			} else {
				w.Header().Set("WWW-Authenticate", "Basic realm=\"GeoNet MTR\"")
				http.Error(w, "Access denied", http.StatusUnauthorized)
				return
			}

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

/*
health does not require auth - for use with AWS EB load balancer checks.
*/
func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

/*
refreshMetics deletes old metrics and udpdates materialized views.
*/
func refreshMetrics() {
	ticker := time.NewTicker(time.Minute * 1).C
	var err error

	for {
		select {
		case <-ticker:
			if _, err = db.Exec(`DELETE FROM field.metric WHERE time < now() - interval '28 days';`); err != nil {
				log.Print(err.Error())
			}
			if err = refreshFieldLatestMetric(); err != nil {
				log.Print(err.Error())
			}
		}
	}
}

func refreshFieldLatestMetric() error {
	_, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY field.metric_summary`)
	return err
}
