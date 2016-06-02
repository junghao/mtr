package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/map180"
	"github.com/GeoNet/mtr/mtrapp"
	"github.com/GeoNet/weft"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"time"
)

var mux *http.ServeMux
var db *sql.DB
var dbR *sql.DB // Database connection with read only credentials
var wm *map180.Map180
var userW = os.Getenv("MTR_USER")
var keyW = os.Getenv("MTR_KEY")

func init() {
	mux = http.NewServeMux()
	mux.HandleFunc("/", weft.MakeHandlerAPI(home))
	mux.HandleFunc("/tag/", weft.MakeHandlerAPI(tagHandler))
	mux.HandleFunc("/tag", weft.MakeHandlerAPI(tagsHandler))
	mux.HandleFunc("/field/model", weft.MakeHandlerAPI(fieldModelHandler))
	mux.HandleFunc("/field/device", weft.MakeHandlerAPI(fieldDeviceHandler))
	mux.HandleFunc("/field/type", weft.MakeHandlerAPI(fieldTypeHandler))
	mux.HandleFunc("/field/metric", weft.MakeHandlerAPI(fieldMetricHandler))
	mux.HandleFunc("/field/metric/summary", weft.MakeHandlerAPI(fieldMetricLatestHandler))
	mux.HandleFunc("/field/metric/threshold", weft.MakeHandlerAPI(fieldThresholdHandler))
	mux.HandleFunc("/field/metric/tag", weft.MakeHandlerAPI(fieldMetricTagHandler))
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/data/site", weft.MakeHandlerAPI(dataSiteHandler))
	mux.HandleFunc("/data/type", weft.MakeHandlerAPI(dataTypeHandler))
	mux.HandleFunc("/data/latency", weft.MakeHandlerAPI(dataLatencyHandler))
	mux.HandleFunc("/data/latency/summary", weft.MakeHandlerAPI(dataLatencySummaryHandler))
	mux.HandleFunc("/data/latency/tag", weft.MakeHandlerAPI(dataLatencyTagHandler))
	mux.HandleFunc("/data/latency/threshold", weft.MakeHandlerAPI(dataLatencyThresholdHandler))
	mux.HandleFunc("/app/metric", weft.MakeHandlerAPI(appMetricHandler))
	mux.HandleFunc("/application/metric", weft.MakeHandlerAPI(applicationMetricHandler))
	mux.HandleFunc("/application/counter", weft.MakeHandlerAPI(applicationCounterHandler))
	mux.HandleFunc("/application/timer", weft.MakeHandlerAPI(applicationTimerHandler))
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

	db.SetMaxIdleConns(30)
	db.SetMaxOpenConns(30)

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

	dbR.SetMaxIdleConns(30)
	dbR.SetMaxOpenConns(30)

	if err = dbR.Ping(); err != nil {
		log.Println("ERROR: problem pinging DB - is it up and contactable? 500s will be served")
	}

	// For map zoom regions other than NZ will need to read some config from somewhere.
	wm, err = map180.Init(dbR, map180.Region(`newzealand`), 256000000)
	if err != nil {
		log.Printf("ERROR: problem with map180 config: %s", err.Error())
	}

	go deleteMetrics()
	go refreshViewsTimed()

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":8080", inbound(mux)))
}

func home(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	return &weft.NotFound
}

// inbound wraps the mux and adds basic auth.
func inbound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT", "DELETE", "POST":
			if user, password, ok := r.BasicAuth(); ok && userW == user && keyW == password {
				h.ServeHTTP(w, r)
			} else {
				http.Error(w, "Access denied", http.StatusUnauthorized)
				mtrapp.StatusUnauthorized.Inc()
				return
			}
		case "GET":
			h.ServeHTTP(w, r)
		default:
			weft.Write(w, r, &weft.MethodNotAllowed)
			weft.MethodNotAllowed.Count()
			return
		}
	})
}

/*
health does not require auth - for use with AWS EB load balancer checks.
*/
func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

// TODO delete app instance and time source that have no metrics?

/*
deleteMetrics deletes old metrics.
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

func refreshViewsTimed() {
	ticker := time.NewTicker(time.Second * 20).C
	for {
		select {
		case <-ticker:
			if err := refreshViews(); err != nil {
				log.Println(err)
			}
		}
	}
}

func refreshViews() error {
	if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY data.latency_summary`); err != nil {
		return err
	}

	if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY field.metric_summary`); err != nil {
		return err
	}

	return nil
}
