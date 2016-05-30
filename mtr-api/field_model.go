package main

import (
	"bytes"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
)

// fieldModel - table field.model
// field devices have a model.
type fieldModel struct {
}

func (a fieldModel) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"modelID"}, []string{}); !res.Ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO field.model(modelID) VALUES($1)`, r.URL.Query().Get("modelID")); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (a fieldModel) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"modelID"}, []string{}); !res.Ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM field.model where modelID = $1`, r.URL.Query().Get("modelID")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a fieldModel) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	var s string

	if err := dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT modelID as "ModelID" FROM field.model) l`).Scan(&s); err != nil {
		return weft.InternalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &weft.StatusOK
}
