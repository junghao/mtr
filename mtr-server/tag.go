package main

import (
	"bytes"
	"database/sql"
	"net/http"
)

func tagHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		var localityID, sourceID, typeID, tag string
		var err error

		if localityID, sourceID, typeID, err = lst(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if tag = r.URL.Query().Get("tag"); tag == "" {
			http.Error(w, "tag is a required parameter", http.StatusBadRequest)
			return
		}

		// Insert and update the tag

		// Ignore errors - will catch any on the update.
		db.Exec(`INSERT INTO field.tag(tag) VALUES($1)`, tag)

		// update is currently a noop to check tag exists.
		var c sql.Result
		if c, err = db.Exec(`UPDATE field.tag SET tag=$1 WHERE tag=$1`, tag); err != nil {
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
			http.Error(w, "no data inserted check tag is valid.", http.StatusBadRequest)
			return
		}

		// Tag the metric
		if c, err = db.Exec(`INSERT INTO field.metric_tag(localityPK, sourcePK, typePK, tagPK) 
			select localityPK, sourcePK, typePK, tagPK 
			FROM field.locality, field.source, field.type, field.tag 
			WHERE 
			localityID = $1
			AND 
			sourceID = $2 
			AND
			typeID = $3
			AND
			tag = $4`,
			localityID, sourceID, typeID, tag); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		i, err = c.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if i == 0 {
			http.Error(w, "no metric tag added check *ID parameters are valid.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "DELETE":
		var localityID, sourceID, typeID, tag string
		var err error

		if localityID, sourceID, typeID, err = lst(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if tag = r.URL.Query().Get("tag"); tag == "" {
			http.Error(w, "tag is a required parameter", http.StatusBadRequest)
			return
		}

		if _, err = db.Exec(`DELETE FROM field.metric_tag USING field.locality, field.source, field.type, field.tag
			WHERE metric_tag.localityPK = locality.localityPK 
			AND metric_tag.sourcePK = source.sourcePK 
			AND metric_tag.typePK = type.typePK 
			AND metric_tag.tagPK = tag.tagPK 
			AND locality.localityID = $1
			AND source.sourceID = $2
			AND type.typeID = $3
			AND tag.tag = $4`, localityID, sourceID, typeID, tag); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	case "GET":
		tag := r.URL.Query().Get("tag") // tag is optional

		var err error
		var b bytes.Buffer

		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			w.Header().Set("Content-Type", "application/json;version=1")
			switch tag {
			case "":
				err = tagsJSONV1(&b)
			default:
				err = metricTagsJSONV1(tag, &b)
			}
		default:
			http.Error(w, "specify accept", http.StatusNotAcceptable)
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

func metricTagsJSONV1(tag string, b *bytes.Buffer) (err error) {
	var d string
	if err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]') 
		FROM (
			SELECT localityID AS "LocalityID",
			sourceID AS "SourceID", 
			typeID AS "TypeID" FROM 
			field.tag JOIN field.metric_tag USING(tagpk) 
			JOIN field.locality USING (localitypk) 
			JOIN field.source USING (sourcepk) 
			JOIN field.type USING (typepk) WHERE tag = $1) l`, tag).Scan(&d); err != nil {
		return
	}

	b.WriteString(d)

	return
}

func tagsJSONV1(b *bytes.Buffer) (err error) {
	var d string
	if err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (select tag AS "Tag" from field.tag order by tag asc) l`).Scan(&d); err != nil {
		return
	}

	b.WriteString(d)

	return
}
