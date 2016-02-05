package main

import (
	"database/sql"
	"net/http"
	"strconv"
)

func localityHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		var localityID, name string
		var latitude, longitude float64
		var err error

		if localityID = r.URL.Query().Get("localityID"); localityID == "" {
			http.Error(w, "localityID is a required parameter", http.StatusBadRequest)
			return
		}

		if name = r.URL.Query().Get("name"); name == "" {
			http.Error(w, "name is a required parameter", http.StatusBadRequest)
			return
		}

		if latitude, err = strconv.ParseFloat(r.URL.Query().Get("latitude"), 64); err != nil {
			http.Error(w, "latitude invalid: "+err.Error(), http.StatusBadRequest)
			return
		}

		if longitude, err = strconv.ParseFloat(r.URL.Query().Get("longitude"), 64); err != nil {
			http.Error(w, "longitude invalid: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Ignore errors - will catch any on the update.
		db.Exec(`INSERT INTO field.locality(localityID, name, latitude, longitude) VALUES($1,$2,$3,$4)`,
			localityID, name, latitude, longitude)

		var c sql.Result
		if c, err = db.Exec(`UPDATE field.locality SET name=$2, latitude=$3, longitude=$4 WHERE localityID=$1`,
			localityID, name, latitude, longitude); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var i int64
		i, err = c.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if i == 0 {
			http.Error(w, "no data inserted check localityID is valid.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "DELETE":
		var localityID string

		if localityID = r.URL.Query().Get("localityID"); localityID == "" {
			http.Error(w, "localityID is a required parameter", http.StatusBadRequest)
			return
		}

		if _, err := db.Exec(`DELETE FROM field.locality WHERE localityID = $1`, localityID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
