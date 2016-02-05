package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
)

var maxAge = time.Duration(-672 * time.Hour)
var future = time.Duration(10 * time.Second)

func fieldMetricHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		var localityID, modelID, typeID, code string
		var t time.Time
		var v int
		var err error

		if localityID = r.URL.Query().Get("localityID"); localityID == "" {
			http.Error(w, "localityID is a required parameter", http.StatusBadRequest)
			return
		}

		if modelID = r.URL.Query().Get("modelID"); modelID == "" {
			http.Error(w, "modelID is a required parameter", http.StatusBadRequest)
			return
		}

		if typeID = r.URL.Query().Get("typeID"); typeID == "" {
			http.Error(w, "typeID is a required parameter", http.StatusBadRequest)
			return
		}

		// code is optional
		if code = r.URL.Query().Get("code"); code == "" {
			code = "NO-CODE"
		}

		if t, err = time.Parse(time.RFC3339, r.URL.Query().Get("time")); err != nil {
			http.Error(w, "invalid time: "+err.Error(), http.StatusBadRequest)
			return
		}

		now := time.Now().UTC()

		if t.Before(now.Add(maxAge)) {
			http.Error(w, "old metric", http.StatusBadRequest)
			return
		}

		if now.Add(future).Before(t) {
			http.Error(w, "future metric", http.StatusBadRequest)
			return
		}

		if v, err = strconv.Atoi(r.URL.Query().Get("value")); err != nil {
			http.Error(w, "invalid value: "+err.Error(), http.StatusBadRequest)
			return
		}

		var c sql.Result
		if c, err = db.Exec(`INSERT INTO field.metric(localityPK, modelPK, sitePK, metricTypePK, time, value) 
			select localityPK, modelPK, sitePK, metricTypePK, $4, $5 
			FROM field.locality, field.model, field.metricType, field.site 
			WHERE 
			localityID = $1
			AND 
			modelID = $2 
			AND
			metricTypeID = $3
			AND 
			code = $6`,
			localityID, modelID, typeID, t, int32(v), code); err != nil {
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
			http.Error(w, "no data inserted check *ID parameters and code (if supplied) are valid.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "DELETE":
		var localityID, modelID string
		var err error

		if localityID = r.URL.Query().Get("localityID"); localityID == "" {
			http.Error(w, "localityID is a required parameter", http.StatusBadRequest)
			return
		}

		if modelID = r.URL.Query().Get("modelID"); modelID == "" {
			http.Error(w, "modelID is a required parameter", http.StatusBadRequest)
			return
		}

		if _, err = db.Exec(`DELETE FROM field.metric USING field.locality, field.model
			WHERE metric.localityPK = locality.localityPK 
			AND metric.modelPK = model.modelPK 
			AND locality.localityID = $1
			AND model.modelID = $2`, localityID, modelID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
