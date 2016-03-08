package main

import (
	"bytes"
	"database/sql"
	"github.com/lib/pq"
	"net/http"
)

type fieldModel struct {
	modelID string
}

func fieldModelPK(modelID string) (int, *result) {
	// TODO - if these don't change they could be app layer cached (for success only).
	var pk int

	if err := dbR.QueryRow(`SELECT modelPK FROM field.model where modelID = $1`, modelID).Scan(&pk); err != nil {
		if err == sql.ErrNoRows {
			return pk, badRequest("unknown modelID")
		}
		return pk, internalServerError(err)
	}

	return pk, &statusOK
}

func (f *fieldModel) save(r *http.Request) *result {
	if res := checkQuery(r, []string{"modelID"}, []string{}); !res.ok {
		return res
	}

	f.modelID = r.URL.Query().Get("modelID")

	if _, err := db.Exec(`INSERT INTO field.model(modelID) VALUES($1)`, f.modelID); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (f *fieldModel) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{"modelID"}, []string{}); !res.ok {
		return res
	}

	f.modelID = r.URL.Query().Get("modelID")

	if _, err := db.Exec(`DELETE FROM field.model where modelID = $1`, f.modelID); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (f *fieldModel) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	var s string

	if err := dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT modelID as "ModelID" FROM field.model) l`).Scan(&s); err != nil {
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &statusOK
}
