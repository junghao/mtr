package main

import (
	"bytes"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
	"strconv"
)

func (a *fieldDevice) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "modelID", "latitude", "longitude"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	a.id = v.Get("deviceID")

	if res := a.fieldModel.read(v.Get("modelID")); !res.Ok {
		return res
	}

	var err error

	if a.latitude, err = strconv.ParseFloat(v.Get("latitude"), 64); err != nil {
		return weft.BadRequest("latitude invalid")
	}

	if a.longitude, err = strconv.ParseFloat(v.Get("longitude"), 64); err != nil {
		return weft.BadRequest("longitude invalid")
	}

	if _, err := db.Exec(`INSERT INTO field.device(deviceID, modelPK, latitude, longitude) VALUES($1, $2, $3, $4)`,
		a.id, a.fieldModel.pk, a.latitude, a.longitude); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
			// TODO should the be an update here?
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (a *fieldDevice) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID"}, []string{}); !res.Ok {
		return res
	}

	a.id = r.URL.Query().Get("deviceID")

	fieldDeviceCache.Lock()
	defer fieldDeviceCache.Unlock()

	if _, err := db.Exec(`DELETE FROM field.device where deviceID = $1`, a.id); err != nil {
		return weft.InternalServerError(err)
	}

	delete(fieldDeviceCache.m, a.id)

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
