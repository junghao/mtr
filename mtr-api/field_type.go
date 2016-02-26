package main

import (
	"bytes"
	"net/http"
)

type fieldType struct{}

func (f *fieldType) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	var s string

	if err := dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT typeID AS "TypeID", description AS "Description", unit AS "Unit"  FROM field.type) l`).Scan(&s); err != nil {
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &statusOK
}
