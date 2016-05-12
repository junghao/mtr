package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
	"strconv"
)

type fieldDevice struct {
	devicePK            int
	deviceID            string
	longitude, latitude float64
}

func (f *fieldDevice) loadPK(r *http.Request) *weft.Result {
	if err := dbR.QueryRow(`SELECT devicePK FROM field.device where deviceID = $1`,
		r.URL.Query().Get("deviceID")).Scan(&f.devicePK); err != nil {
		if err == sql.ErrNoRows {
			return weft.BadRequest("unknown deviceID")
		}
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

// TODO deprecated get rid of this.
func fieldDevicePK(deviceID string) (int, *weft.Result) {
	// TODO - if these don't change they could be app layer cached (for success only).
	var pk int

	if err := dbR.QueryRow(`SELECT devicePK FROM field.device where deviceID = $1`, deviceID).Scan(&pk); err != nil {
		if err == sql.ErrNoRows {
			return pk, weft.BadRequest("unknown deviceID")
		}
		return pk, weft.InternalServerError(err)
	}

	return pk, &weft.StatusOK
}

func (f *fieldDevice) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "modelID", "latitude", "longitude"}, []string{}); !res.Ok {
		return res
	}

	f.deviceID = r.URL.Query().Get("deviceID")

	var res *weft.Result
	var modelPK int

	if modelPK, res = fieldModelPK(r.URL.Query().Get("modelID")); !res.Ok {
		return res
	}

	var err error

	if f.latitude, err = strconv.ParseFloat(r.URL.Query().Get("latitude"), 64); err != nil {
		return weft.BadRequest("latitude invalid")
	}

	if f.longitude, err = strconv.ParseFloat(r.URL.Query().Get("longitude"), 64); err != nil {
		return weft.BadRequest("longitude invalid")
	}

	if _, err := db.Exec(`INSERT INTO field.device(deviceID, modelPK, latitude, longitude) VALUES($1, $2, $3, $4)`,
		f.deviceID, modelPK, f.latitude, f.longitude); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
			// TODO should the be an update here?
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (f *fieldDevice) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID"}, []string{}); !res.Ok {
		return res
	}

	f.deviceID = r.URL.Query().Get("deviceID")

	if _, err := db.Exec(`DELETE FROM field.device where deviceID = $1`, f.deviceID); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (f *fieldDevice) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	var s string

	if err := dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT deviceid AS "DeviceID", modelid AS "ModelID", latitude AS "Latitude",
			longitude AS "Longitude" FROM field.device JOIN field.model USING(modelpk)) l`).Scan(&s); err != nil {
		return weft.InternalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &weft.StatusOK
}
