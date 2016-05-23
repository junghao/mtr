package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"
	"net/http"
	"strconv"
)

func (f *fieldThreshold) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID", "lower", "upper"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()
	var err error

	if f.lower, err = strconv.Atoi(v.Get("lower")); err != nil {
		return weft.BadRequest("invalid lower")
	}

	if f.upper, err = strconv.Atoi(v.Get("upper")); err != nil {
		return weft.BadRequest("invalid upper")
	}

	if res := f.fieldDeviceType.read(v.Get("deviceID"), v.Get("typeID")); !res.Ok {
		return res
	}

	if _, err = db.Exec(`INSERT INTO field.threshold(devicePK, typePK, lower, upper) 
		VALUES ($1,$2,$3,$4)`,
		f.fieldDevice.pk, f.fieldType.pk, f.lower, f.upper); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return weft.InternalServerError(err)
		}
	}

	if _, err = db.Exec(`UPDATE field.threshold SET lower=$3, upper=$4 
		WHERE 
		devicePK = $1 
		AND
		typePK = $2`,
		f.fieldDevice.pk, f.fieldType.pk, f.lower, f.upper); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (f *fieldThreshold) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{}); !res.Ok {
		return res
	}

	var err error

	v := r.URL.Query()

	if res := f.fieldDeviceType.read(v.Get("deviceID"), v.Get("typeID")); !res.Ok {
		return res
	}

	if _, err = db.Exec(`DELETE FROM field.threshold 
		WHERE devicePK = $1
		AND typePK = $2 `,
		f.fieldDevice.pk, f.fieldType.pk); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (f *fieldThreshold) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	var s string

	if err := dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]')
		FROM (SELECT deviceID as "DeviceID", typeID as "TypeID", 
		lower as "Lower", upper AS "Upper" 
		FROM 
		field.threshold JOIN field.device USING (devicepk) 
		JOIN field.type USING (typepk)) l`).Scan(&s); err != nil {
		return weft.InternalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &weft.StatusOK
}

func (f *fieldThreshold) proto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

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

	h.Set("Content-Type", "application/x-protobuf")

	return &weft.StatusOK
}
