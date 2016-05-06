package main

import (
	"database/sql"
	"github.com/GeoNet/map180"
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
var wm *map180.Map180

var eol = []byte("\n")

func init() {
	userW = os.Getenv("MTR_USER")
	keyW = os.Getenv("MTR_KEY")
	userR = os.Getenv("MTR_USER_R")
	keyR = os.Getenv("MTR_KEY_R")

	mux = http.NewServeMux()
	mux.HandleFunc("/tag", toHandler(tagHandler))
	mux.HandleFunc("/app/metric", toHandler(appMetricHandler))
	mux.HandleFunc("/field/model", toHandler(fieldModelHandler))
	mux.HandleFunc("/field/device", toHandler(fieldDeviceHandler))
	//mux.HandleFunc("/field/tag", toHandler(fieldTagHandler))
	mux.HandleFunc("/field/type", toHandler(fieldTypeHandler))
	mux.HandleFunc("/field/metric", toHandler(fieldMetricHandler))
	mux.HandleFunc("/field/metric/summary", toHandler(fieldMetricLatestHandler))
	mux.HandleFunc("/field/metric/threshold", toHandler(fieldThresholdHandler))
	mux.HandleFunc("/field/metric/tag", toHandler(fieldMetricTagHandler))
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/data/site", toHandler(dataSiteHandler))
	mux.HandleFunc("/data/latency", toHandler(dataLatencyHandler))
	mux.HandleFunc("/data/latency/summary", toHandler(dataLatencySummaryHandler))
	mux.HandleFunc("/data/latency/tag", toHandler(dataLatencyTagHandler))
	mux.HandleFunc("/data/latency/threshold", toHandler(dataLatencyThresholdHandler))
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

	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

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

	// For map zoom regions other than NZ will need to read some config from somewhere.
	wm, err = map180.Init(dbR, map180.Region(`newzealand`), 256000000)
	if err != nil {
		log.Println("ERROR: problem with map180 config: %s", err)
	}

	go deleteMetrics()
	go refreshViews()

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

/*
health does not require auth - for use with AWS EB load balancer checks.
*/
func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

// TODO delete app instance and time source that have no metrics?

/*
deleteMetics deletes old metrics.
*/
func deleteMetrics() {
	ticker := time.NewTicker(time.Minute).C
	var err error
	for {
		select {
		case <-ticker:
			if _, err = db.Exec(`DELETE FROM field.metric WHERE time < now() - interval '40 days'`); err != nil {
				log.Println(err)
			}

			if _, err = db.Exec(`DELETE FROM data.latency WHERE time < now() - interval '40 days'`); err != nil {
				log.Println(err)
			}

			if _, err = db.Exec(`DELETE FROM app.metric WHERE time < now() - interval '28 days'`); err != nil {
				log.Println(err)
			}

			if _, err = db.Exec(`DELETE FROM app.counter WHERE time < now() - interval '28 days'`); err != nil {
				log.Println(err)
			}

			if _, err = db.Exec(`DELETE FROM app.timer WHERE time < now() - interval '28 days'`); err != nil {
				log.Println(err)
			}
		}
	}
}

func refreshViews() {
	ticker := time.NewTicker(time.Second * 20).C
	var err error
	for {
		select {
		case <-ticker:
			if _, err = db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY data.latency_summary`); err != nil {
				log.Println(err)
			}

			if _, err = db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY field.metric_summary`); err != nil {
				log.Println(err)
			}
		}
	}
}
