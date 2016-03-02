package main

import (
	"bytes"
	"database/sql"
	"github.com/lib/pq"
	"net/http"
)

type fieldTag struct {
	tag string
}

func (f *fieldTag) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{"tag"}); !res.ok {
		return res
	}

	f.tag = r.URL.Query().Get("tag")

	var s string
	var err error

	switch f.tag {
	case "":
		err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (select tag AS "Tag" from field.tag order by tag asc) l`).Scan(&s)
	default:
		err = dbR.QueryRow(`SELECT row_to_json(l) 
		FROM (
			SELECT localityID AS "LocalityID",
			deviceID AS "DeviceID", 
			typeID AS "TypeID" FROM 
			field.tag JOIN field.metric_tag USING(tagpk) 
			JOIN field.locality USING (localitypk) 
			JOIN field.device USING (devicepk) 
			JOIN field.type USING (typepk) WHERE tag = $1) l`, f.tag).Scan(&s)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return &notFound
		}
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &statusOK
}

func (f *fieldTag) save(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"localityID", "deviceID", "typeID", "tag"}, []string{}); !res.ok {
		return res
	}

	f.tag = r.URL.Query().Get("tag")

	var fm fieldMetric

	if res := fm.loadID(r); !res.ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO field.tag(tag) VALUES($1)`, f.tag); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == `23505` {
			// ignore unique constraint errors
		} else {
			return internalServerError(err)
		}
	}

	// Tag the metric
	if _, err := db.Exec(`INSERT INTO field.metric_tag(localityPK, devicePK, typePK, tagPK) 
			SELECT $1, $2, $3, tagPK 
			FROM field.tag WHERE tag = $4`,
		fm.localityPK, fm.devicePK, fm.typePK, f.tag); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == `23505` {
			// ignore unique constraint errors
		} else {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (f *fieldTag) delete(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"localityID", "deviceID", "typeID", "tag"}, []string{}); !res.ok {
		return res
	}

	f.tag = r.URL.Query().Get("tag")

	var fm fieldMetric

	if res := fm.loadID(r); !res.ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM field.metric_tag USING field.tag
			WHERE localityPK = $1
			AND devicePK = $2
			AND typePK = $3
			AND metric_tag.tagPK = tag.tagPK
			AND tag.tag = $4`, fm.localityPK, fm.devicePK, fm.typePK, f.tag); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}
