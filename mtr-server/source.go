package main

import (
	"bytes"
	"database/sql"
	"net/http"
)

func sourceHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		var sourceID string
		var err error

		if sourceID = r.URL.Query().Get("sourceID"); sourceID == "" {
			http.Error(w, "sourceID is a required parameter", http.StatusBadRequest)
			return
		}

		// Ignore errors - will catch any on the update.
		db.Exec(`INSERT INTO field.source(sourceID) VALUES($1)`, sourceID)

		// update is currently a noop to check sourceID exists.
		var c sql.Result
		if c, err = db.Exec(`UPDATE field.source SET sourceID=$1 WHERE sourceID=$1`, sourceID); err != nil {
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
			http.Error(w, "no data inserted check sourceID is valid.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "DELETE":
		var sourceID string
		var err error

		if sourceID = r.URL.Query().Get("sourceID"); sourceID == "" {
			http.Error(w, "sourceID is a required parameter", http.StatusBadRequest)
			return
		}

		if _, err = db.Exec(`DELETE FROM field.source where sourceID = $1`, sourceID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	case "GET":
		var err error
		var b bytes.Buffer

		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			w.Header().Set("Content-Type", "application/json;version=1")
			err = sourcesJSONV1(&b)
		default:
			http.Error(w, "specify accept", http.StatusNotAcceptable)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b.WriteTo(w)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func sourcesJSONV1(b *bytes.Buffer) (err error) {
	var s string

	if err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT sourceID as "SourceID" FROM field.Source) l`).Scan(&s); err != nil {
		return
	}

	b.WriteString(s)

	return
}
