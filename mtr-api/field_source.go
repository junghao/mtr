package main

import (
	"bytes"
	"github.com/lib/pq"
	"net/http"
)

type fieldSource struct {
	sourceID string
}

func (f *fieldSource) save(r *http.Request) *result {
	if res := checkQuery(r, []string{"sourceID"}, []string{}); !res.ok {
		return res
	}

	f.sourceID = r.URL.Query().Get("sourceID")

	if _, err := db.Exec(`INSERT INTO field.source(sourceID) VALUES($1)`, f.sourceID); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == `23505` {
			// ignore unique constraint errors
		} else {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (f *fieldSource) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{"sourceID"}, []string{}); !res.ok {
		return res
	}

	f.sourceID = r.URL.Query().Get("sourceID")

	if _, err := db.Exec(`DELETE FROM field.source where sourceID = $1`, f.sourceID); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (f *fieldSource) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	var s string

	if err := dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT sourceID as "SourceID" FROM field.Source) l`).Scan(&s); err != nil {
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &statusOK
}
