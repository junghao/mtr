package main

import (
	"database/sql"
	"net/http"
)

func modelHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		var modelID string
		var err error

		if modelID = r.URL.Query().Get("modelID"); modelID == "" {
			http.Error(w, "modelID is a required parameter", http.StatusBadRequest)
			return
		}

		// Ignore errors - will catch any on the update.
		db.Exec(`INSERT INTO field.model(modelID) VALUES($1)`, modelID)

		// update is currently a noop to check modelID exists.
		var c sql.Result
		if c, err = db.Exec(`UPDATE field.model SET modelID=$1 WHERE modelID=$1`, modelID); err != nil {
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
			http.Error(w, "no data inserted check modelID is valid.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "DELETE":
		var modelID string
		var err error

		if modelID = r.URL.Query().Get("modelID"); modelID == "" {
			http.Error(w, "modelID is a required parameter", http.StatusBadRequest)
			return
		}

		if _, err = db.Exec(`DELETE FROM field.model where modelID = $1`, modelID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
