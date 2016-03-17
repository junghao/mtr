package main

import (
	"bytes"
	"database/sql"
	"net/http"
)

type fieldTag struct {
	tag string
}

func (f *fieldTag) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	var s string

	if err := dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT tag as "Tag" FROM  field.tag) l`).Scan(&s); err != nil {
		if err == sql.ErrNoRows {
			return &notFound
		}
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &statusOK
}
