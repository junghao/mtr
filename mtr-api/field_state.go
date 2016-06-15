package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"net/http"
	"strconv"
	"time"
)

type fieldState struct {
}

func (f fieldState) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID", "time", "value"}, []string{}); !res.Ok {
		return res
	}

	q := r.URL.Query()
	deviceID := q.Get("deviceID")
	typeID := q.Get("typeID")

	var err error
	var value bool
	if value, err = strconv.ParseBool(q.Get("value")); err != nil {
		return weft.BadRequest("invalid value")
	}

	var t time.Time
	if t, err = time.Parse(time.RFC3339, q.Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	var result sql.Result
	if result, err = db.Exec(`UPDATE field.state SET
				time = $3, value = $4
				WHERE devicePK = (SELECT devicePK from field.device WHERE deviceID = $1)
				AND typePK = (SELECT typePK from field.state_type WHERE typeID = $2)`,
		deviceID, typeID, t, value); err != nil {
		return weft.InternalServerError(err)
	}

	// If no rows change either the values are old or it's the first time we've seen this metric.
	var u int64
	if u, err = result.RowsAffected(); err != nil {
		return weft.InternalServerError(err)
	}

	if u == 1 {
		return &weft.StatusOK
	} else if result, err = db.Exec(`INSERT INTO field.state(devicePK, typePK, time, value)
					SELECT devicePK, typePK, $3, $4
					FROM field.device, field.state_type
					WHERE deviceID = $1
					AND typeID = $2`,
		deviceID, typeID, t, value); err == nil {

		var i int64
		if i, err = result.RowsAffected(); err != nil {
			return weft.InternalServerError(err)
		}
		if i == 1 {
			return &weft.StatusOK
		}
	}

	return weft.InternalServerError(err)
}

func (f fieldState) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{}); !res.Ok {
		return res
	}

	q := r.URL.Query()

	if _, err := db.Exec(`DELETE FROM field.state
			WHERE devicePK = (SELECT devicePK FROM field.device WHERE deviceID = $1)
			AND typePK = (SELECT typePK from field.state_type WHERE typeID = $2)`,
		q.Get("deviceID"), q.Get("typeID")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (f fieldState) allProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT deviceID, typeID, time, value
				FROM field.state
				JOIN field.device USING (devicePK)
				JOIN field.state_type USING (typePK)`); err != nil {
		return weft.InternalServerError(err)
	}

	var fr mtrpb.FieldStateResult
	var t time.Time

	for rows.Next() {
		var s mtrpb.FieldState

		if err = rows.Scan(&s.DeviceID, &s.TypeID, &t, &s.Value); err != nil {
			return weft.InternalServerError(err)
		}

		// Convert from Go's time.Time to unix seconds since epoch discarding nanosecs
		s.Seconds = t.Unix()

		fr.Result = append(fr.Result, &s)
	}

	var by []byte
	if by, err = proto.Marshal(&fr); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	h.Set("Content-Type", "application/x-protobuf")

	return &weft.StatusOK
}
