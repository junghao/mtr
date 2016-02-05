package main

import (
	"database/sql"
	"net/http"
)

func siteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		var code string
		var err error

		if code = r.URL.Query().Get("code"); code == "" {
			http.Error(w, "code is a required parameter", http.StatusBadRequest)
			return
		}

		// Ignore errors - will catch any on the update.
		db.Exec(`INSERT INTO field.site(code) VALUES($1)`, code)

		// update is currently a noop to check code exists.
		var c sql.Result
		if c, err = db.Exec(`UPDATE field.site SET code=$1 WHERE code=$1`, code); err != nil {
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
			http.Error(w, "no data inserted check code is valid.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "DELETE":
		var code string
		var err error

		if code = r.URL.Query().Get("code"); code == "" {
			http.Error(w, "code is a required parameter", http.StatusBadRequest)
			return
		}

		if _, err = db.Exec(`DELETE from field.site where code = $1`, code); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
