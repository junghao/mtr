package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"
	"net/http"
)

func fieldStateTagPut(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	var err error
	var result sql.Result

	if result, err = db.Exec(`INSERT INTO field.state_tag(devicePK, typePK, tagPK)
				SELECT devicePK, typePK, tagPK
				FROM field.device, field.state_type, mtr.tag
				WHERE deviceID = $1
				AND typeID = $2
				AND tag = $3`,
		v.Get("deviceID"), v.Get("typeID"), v.Get("tag")); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
			return &weft.StatusOK
		} else {
			return weft.InternalServerError(err)
		}
	}

	var i int64
	if i, err = result.RowsAffected(); err != nil {
		return weft.InternalServerError(err)
	}
	if i != 1 {
		return weft.BadRequest("Didn't create row, check your query parameters exist")
	}

	return &weft.StatusOK
}

func fieldStateTagDelete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	if _, err := db.Exec(`DELETE FROM field.state_tag
			WHERE devicePK = (SELECT devicePK FROM field.device WHERE deviceID = $1)
			AND typePK = (SELECT typePK FROM field.state_type WHERE typeID = $2)
			AND tagPK = (SELECT tagPK FROM mtr.tag WHERE tag = $3)`,
		v.Get("deviceID"), v.Get("typeID"), v.Get("tag")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func fieldStateTagProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT deviceID, tag, typeID from field.state_tag
				JOIN mtr.tag USING (tagpk)
				JOIN field.device USING (devicepk)
				JOIN field.state_type USING (typepk)
				ORDER BY tag ASC`); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	var ts mtrpb.FieldStateTagResult

	for rows.Next() {
		var t mtrpb.FieldStateTag

		if err = rows.Scan(&t.DeviceID, &t.Tag, &t.TypeID); err != nil {
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
