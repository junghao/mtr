package main

import (
	"bytes"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
)

func (a *fieldModel) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"modelID"}, []string{}); !res.Ok {
		return res
	}

	a.id = r.URL.Query().Get("modelID")

	if _, err := db.Exec(`INSERT INTO field.model(modelID) VALUES($1)`, a.id); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (a *fieldModel) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"modelID"}, []string{}); !res.Ok {
		return res
	}

	a.id = r.URL.Query().Get("modelID")

	// Deleting a model can remove all the devices for that model so reset the field device cache.
	fieldDeviceCache.Lock()
	defer fieldDeviceCache.Unlock()

	fieldDeviceCache.m = make(map[string]int)

	if _, err := db.Exec(`DELETE FROM field.model where modelID = $1`, a.id); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a *fieldModel) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
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
