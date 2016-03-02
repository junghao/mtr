package main

import (
	"bytes"
	"database/sql"
	"net/http"
)

type fieldType struct{}

func fieldTypePK(typeID string) (int, *result) {
	// TODO - if these don't change they could be app layer cached (for success only).
	var pk int

	if err := dbR.QueryRow(`SELECT typePK FROM field.type where typeID = $1`, typeID).Scan(&pk); err != nil {
		if err == sql.ErrNoRows {
			return pk, badRequest("unknown typeID")
		}
		return pk, internalServerError(err)
	}

	return pk, &statusOK
}

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
