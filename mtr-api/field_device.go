package main

import (
	"bytes"
	"database/sql"
	"github.com/lib/pq"
	"net/http"
	"strconv"
)

type fieldDevice struct {
	devicePK int
	deviceID            string
	longitude, latitude float64
}

func (f *fieldDevice) loadPK(r *http.Request) *result {
	if err := dbR.QueryRow(`SELECT devicePK FROM field.device where deviceID = $1`,
		r.URL.Query().Get("deviceID")).Scan(&f.devicePK); err != nil {
		if err == sql.ErrNoRows {
			return badRequest("unknown deviceID")
		}
		return internalServerError(err)
	}

	return &statusOK
}

// TODO deprecated get rid of this.
func fieldDevicePK(deviceID string) (int, *result) {
	// TODO - if these don't change they could be app layer cached (for success only).
	var pk int

	if err := dbR.QueryRow(`SELECT devicePK FROM field.device where deviceID = $1`, deviceID).Scan(&pk); err != nil {
		if err == sql.ErrNoRows {
			return pk, badRequest("unknown deviceID")
		}
		return pk, internalServerError(err)
	}

	return pk, &statusOK
}

func (f *fieldDevice) save(r *http.Request) *result {
	if res := checkQuery(r, []string{"deviceID", "modelID", "latitude", "longitude"}, []string{}); !res.ok {
		return res
	}

	f.deviceID = r.URL.Query().Get("deviceID")

	var res *result
	var modelPK int

	if modelPK, res = fieldModelPK(r.URL.Query().Get("modelID")); !res.ok {
		return res
	}

	var err error

	if f.latitude, err = strconv.ParseFloat(r.URL.Query().Get("latitude"), 64); err != nil {
		return badRequest("latitude invalid")
	}

	if f.longitude, err = strconv.ParseFloat(r.URL.Query().Get("longitude"), 64); err != nil {
		return badRequest("longitude invalid")
	}

	if _, err := db.Exec(`INSERT INTO field.device(deviceID, modelPK, latitude, longitude) VALUES($1, $2, $3, $4)`,
		f.deviceID, modelPK, f.latitude, f.longitude); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
			// TODO should the be an update here?
		} else {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (f *fieldDevice) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{"deviceID"}, []string{}); !res.ok {
		return res
	}

	f.deviceID = r.URL.Query().Get("deviceID")

	if _, err := db.Exec(`DELETE FROM field.device where deviceID = $1`, f.deviceID); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (f *fieldDevice) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	var s string

	if err := dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT deviceid AS "DeviceID", modelid AS "ModelID", latitude AS "Latitude",
			longitude AS "Longitude" FROM field.device JOIN field.model USING(modelpk)) l`).Scan(&s); err != nil {
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &statusOK
}
