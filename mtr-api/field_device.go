package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"
	"net/http"
	"strconv"
)

func fieldDevicePut(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	var err error
	var latitude, longitude float64

	if latitude, err = strconv.ParseFloat(v.Get("latitude"), 64); err != nil {
		return weft.BadRequest("latitude invalid")
	}

	if longitude, err = strconv.ParseFloat(v.Get("longitude"), 64); err != nil {
		return weft.BadRequest("longitude invalid")
	}

	var result sql.Result

	// TODO - use upsert with PG 9.5?

	// return if insert succeeds
	if result, err = db.Exec(`INSERT INTO field.device(deviceID, modelPK, latitude, longitude)
				SELECT $1, modelPK, $3, $4
				FROM field.model
				WHERE modelID = $2`,
		v.Get("deviceID"), v.Get("modelID"), latitude, longitude); err == nil {
		var i int64
		if i, err = result.RowsAffected(); err != nil {
			return weft.InternalServerError(err)
		}
		if i == 1 {
			return &weft.StatusOK
		}
	}

	// return if update one row
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == errorUniqueViolation {
		if result, err = db.Exec(`UPDATE field.device
					SET latitude = $2, longitude = $3
					WHERE deviceID = $1`,
			v.Get("deviceID"), latitude, longitude); err == nil {
			var i int64
			if i, err = result.RowsAffected(); err != nil {
				return weft.InternalServerError(err)
			}
			if i == 1 {
				return &weft.StatusOK
			}
		}
	}

	if err == nil {
		err = fmt.Errorf("no rows affected, check your query.")
	}

	return weft.InternalServerError(err)
}

func fieldDeviceDelete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if _, err := db.Exec(`DELETE FROM field.device where deviceID = $1`, r.URL.Query().Get("deviceID")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func fieldDeviceProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT deviceid, modelid, latitude, longitude
		FROM
		field.device JOIN field.model USING(modelpk)`); err != nil {
		return weft.InternalServerError(err)
	}

	var fdr mtrpb.FieldDeviceResult

	for rows.Next() {
		var d mtrpb.FieldDevice

		if err = rows.Scan(&d.DeviceID, &d.ModelID, &d.Latitude, &d.Longitude); err != nil {
			return weft.InternalServerError(err)
		}

		fdr.Result = append(fdr.Result, &d)
	}

	var by []byte
	if by, err = proto.Marshal(&fdr); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	return &weft.StatusOK
}
