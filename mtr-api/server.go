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

// the handler wiring and majority of mux routing is generated from weft.toml
// it is created with weftgenapi  It is only necessary to run weftgenapi when
// weft.toml is changed.
//
// go install github.com/GeoNet/weft/weftgenapi
// go generate && gofmt -s -w .
//
//go:generate weftgenapi

var db *sql.DB
var dbR *sql.DB // Database connection with read only credentials
var wm *map180.Map180
var userW = os.Getenv("MTR_USER")
var keyW = os.Getenv("MTR_KEY")

func init() {
	mux.HandleFunc("/", weft.MakeHandlerAPI(home))
	mux.HandleFunc("/health", health)

	// routes for balancers and probes.
	mux.HandleFunc("/soh/up", http.HandlerFunc(up))
	mux.HandleFunc("/soh", http.HandlerFunc(soh))
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

			if _, err = db.Exec(`DELETE FROM field.metric_summary WHERE time < now() - interval '40 days'`); err != nil {
				log.Println(err)
			}

			if _, err = db.Exec(`DELETE FROM data.latency WHERE time < now() - interval '40 days'`); err != nil {
				log.Println(err)
			}

			if _, err = db.Exec(`DELETE FROM data.latency_summary WHERE time < now() - interval '40 days'`); err != nil {
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

// up is for testing that the app has started e.g., for with load balancers.
// It indicates the app is started.  It may still be serving errors.
// Not useful for inclusion in app metrics so weft not used.
func up(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		w.Header().Set("Surrogate-Control", "max-age=86400")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Surrogate-Control", "max-age=10")

	w.Write([]byte("<html><head></head><body>up</body></html>"))
	log.Print("up ok")
}

// soh is for external service probes.
// writes a service unavailable error to w if the service is not working.
// Not useful for inclusion in app metrics so weft not used.
func soh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		w.Header().Set("Surrogate-Control", "max-age=86400")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var c int

	if err := db.QueryRow("SELECT 1").Scan(&c); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("<html><head></head><body>service error</body></html>"))
		log.Printf("ERROR: soh service error %s", err)
		return
	}

	w.Header().Set("Surrogate-Control", "max-age=10")

	w.Write([]byte("<html><head></head><body>ok</body></html>"))
	log.Print("soh ok")
}
