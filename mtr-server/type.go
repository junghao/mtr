package main

import (
	"bytes"
	"net/http"
)

func typeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var err error
		var b bytes.Buffer

		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			w.Header().Set("Content-Type", "application/json;version=1")
			err = typesJSONV1(&b)
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

func typesJSONV1(b *bytes.Buffer) (err error) {
	var s string

	if err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT typeID AS "TypeID", description AS "Description", unit AS "Unit"  FROM field.type) l`).Scan(&s); err != nil {
		return
	}

	b.WriteString(s)

	return
}
