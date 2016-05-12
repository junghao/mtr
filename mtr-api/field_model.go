package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
)

type fieldModel struct {
	modelID string
}

func fieldModelPK(modelID string) (int, *weft.Result) {
	// TODO - if these don't change they could be app layer cached (for success only).
	var pk int

	if err := dbR.QueryRow(`SELECT modelPK FROM field.model where modelID = $1`, modelID).Scan(&pk); err != nil {
		if err == sql.ErrNoRows {
			return pk, weft.BadRequest("unknown modelID")
		}
		return pk, weft.InternalServerError(err)
	}

	return pk, &weft.StatusOK
}

func (f *fieldModel) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"modelID"}, []string{}); !res.Ok {
		return res
	}

	f.modelID = r.URL.Query().Get("modelID")

	if _, err := db.Exec(`INSERT INTO field.model(modelID) VALUES($1)`, f.modelID); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (f *fieldModel) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"modelID"}, []string{}); !res.Ok {
		return res
	}

	f.modelID = r.URL.Query().Get("modelID")

	if _, err := db.Exec(`DELETE FROM field.model where modelID = $1`, f.modelID); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (f *fieldModel) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
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
