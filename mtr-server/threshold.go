package main

import (
	"bytes"
	"database/sql"
	"net/http"
	"strconv"
)

func thresholdHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		var localityID, sourceID, typeID string
		var min, max int
		var err error

		if localityID, sourceID, typeID, err = lst(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if min, err = strconv.Atoi(r.URL.Query().Get("min")); err != nil {
			http.Error(w, "invalid min: "+err.Error(), http.StatusBadRequest)
			return
		}

		if max, err = strconv.Atoi(r.URL.Query().Get("max")); err != nil {
			http.Error(w, "invalid max: "+err.Error(), http.StatusBadRequest)
			return
		}

		// ignore errors, catch them on the update.
		db.Exec(`INSERT INTO field.threshold(localityPK, sourcePK, typePK, min, max) 
			select localityPK, sourcePK, typePK, $4, $5 
			FROM field.locality, field.source, field.type 
			WHERE 
			localityID = $1
			AND 
			sourceID = $2 
			AND
			typeID = $3`,
			localityID, sourceID, typeID, min, max)

		var c sql.Result
		if c, err = db.Exec(`UPDATE field.threshold SET min=$4, max=$5 
			FROM field.locality, field.source, field.type 
			WHERE 
			localityID = $1
			AND 
			sourceID = $2 
			AND
			typeID = $3`,
			localityID, sourceID, typeID, min, max); err != nil {
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
		var localityID, sourceID, typeID string
		var err error

		if localityID, sourceID, typeID, err = lst(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err = db.Exec(`DELETE FROM field.threshold USING field.locality, field.source, field.type
			WHERE threshold.localityPK = locality.localityPK 
			AND threshold.sourcePK = source.sourcePK 
			AND threshold.typePK = type.typePK 
			AND locality.localityID = $1
			AND source.sourceID = $2
			AND type.typeID = $3`, localityID, sourceID, typeID); err != nil {
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
			err = thresholdsJSONV1(&b)
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

func thresholdsJSONV1(b *bytes.Buffer) (err error) {
	var s string

	if err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT localityID as "LocalityID",sourceID as "SourceID", typeID as "TypeID", 
		min as Min, max AS "Max" FROM field.threshold JOIN field.locality USING (localitypk) 
		JOIN field.source USING (sourcepk) JOIN field.type USING (typepk)) l`).Scan(&s); err != nil {
		return
	}

	b.WriteString(s)

	return
}
