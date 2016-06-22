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

// fieldMetricTag - table field.metric_tag
type fieldMetricTag struct {
}

func (f fieldMetricTag) put(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID", "tag"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error
	var result sql.Result

	if result, err = db.Exec(`INSERT INTO field.metric_tag(devicePK, typePK, tagPK)
				SELECT devicePK, typePK, tagPK
				FROM field.device, field.type, mtr.tag
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

func (f fieldMetricTag) delete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID", "tag"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	if _, err := db.Exec(`DELETE FROM field.metric_tag
			WHERE devicePK = (SELECT devicePK FROM field.device WHERE deviceID = $1)
			AND typePK = (SELECT typePK FROM field.type WHERE typeID = $2)
			AND tagPK = (SELECT tagPK FROM mtr.tag WHERE tag = $3)`,
		v.Get("deviceID"), v.Get("typeID"), v.Get("tag")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (t fieldMetricTag) all(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{"deviceID", "typeID"}); !res.Ok {
		return res
	}

	var err error
	var rows *sql.Rows

	deviceID := r.URL.Query().Get("deviceID")
	typeID := r.URL.Query().Get("typeID")

	if deviceID == "" && typeID == "" {
		if rows, err = dbR.Query(`SELECT deviceID, tag, typeID from field.metric_tag
				JOIN mtr.tag USING (tagpk)
				JOIN field.device USING (devicepk)
				JOIN field.type USING (typepk)
				ORDER BY tag ASC`); err != nil {
			return weft.InternalServerError(err)
		}
		defer rows.Close()
	} else if deviceID != "" && typeID != "" {
		if rows, err = dbR.Query(`SELECT deviceID, tag, typeID from field.metric_tag
				JOIN mtr.tag USING (tagpk)
				JOIN field.device USING (devicepk)
				JOIN field.type USING (typepk)
				WHERE deviceID=$1 AND typeID=$2
				ORDER BY tag ASC`, deviceID, typeID); err != nil {
			return weft.InternalServerError(err)
		}
		defer rows.Close()

	} else {
		return weft.BadRequest("Invalid parameter. Please specify both deviceID and typeID.")
	}

	var ts mtrpb.FieldMetricTagResult

	for rows.Next() {
		var t mtrpb.FieldMetricTag

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

	h.Set("Content-Type", "application/x-protobuf")

	return &weft.StatusOK
}
