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

func fieldThresholdPut(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()
	var err error

	var lower, upper int

	if lower, err = strconv.Atoi(v.Get("lower")); err != nil {
		return weft.BadRequest("invalid lower")
	}

	if upper, err = strconv.Atoi(v.Get("upper")); err != nil {
		return weft.BadRequest("invalid upper")
	}

	deviceID := v.Get("deviceID")
	typeID := v.Get("typeID")

	var result sql.Result

	// TODO - use upsert with PG 9.5?

	// return if insert succeeds
	if result, err = db.Exec(`INSERT INTO field.threshold(devicePK, typePK, lower, upper)
		SELECT devicePK, typePK, $3, $4
				FROM field.device, field.type
				WHERE deviceID = $1
				AND typeID = $2`,
		deviceID, typeID, lower, upper); err == nil {
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
		if result, err = db.Exec(`UPDATE field.threshold SET lower=$3, upper=$4
		WHERE devicePK = (SELECT devicePK FROM field.device WHERE deviceID = $1)
		AND typePK = (SELECT typePK FROM field.type WHERE typeID = $2)`,
			deviceID, typeID, lower, upper); err == nil {
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

func fieldThresholdDelete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error

	v := r.URL.Query()

	if _, err = db.Exec(`DELETE FROM field.threshold
		WHERE devicePK = (SELECT devicePK FROM field.device WHERE deviceID = $1)
		AND typePK = (SELECT typePK FROM field.type WHERE typeID = $2)`,
		v.Get("deviceID"), v.Get("typeID")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func fieldThresholdProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT deviceID, typeID, lower, upper
		FROM
		field.threshold JOIN field.device USING (devicepk)
		JOIN field.type USING (typepk)`); err != nil {
		return weft.InternalServerError(err)
	}

	var ts mtrpb.FieldMetricThresholdResult

	for rows.Next() {
		var t mtrpb.FieldMetricThreshold

		if err = rows.Scan(&t.DeviceID, &t.TypeID, &t.Lower, &t.Upper); err != nil {
			return weft.InternalServerError(err)
		}

		ts.Result = append(ts.Result, &t)
	}

	var by []byte
	if by, err = proto.Marshal(&ts); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	return &weft.StatusOK
}
