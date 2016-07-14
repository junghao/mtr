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

func dataLatencyThresholdPut(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()
	var err error

	var lower, upper int

	if lower, err = strconv.Atoi(v.Get("lower")); err != nil {
		return weft.BadRequest("invalid lower")
	}

	if upper, err = strconv.Atoi(v.Get("upper")); err != nil {
		return weft.BadRequest("invalid upper")
	}

	siteID := v.Get("siteID")
	typeID := v.Get("typeID")

	var result sql.Result

	// TODO Change to upsert 9.5

	// return if insert succeeds
	if result, err = db.Exec(`INSERT INTO data.latency_threshold(sitePK, typePK, lower, upper)
				SELECT sitePK, typePK, $3, $4
				FROM data.site, data.type
				WHERE siteID = $1
				AND typeID = $2`,
		siteID, typeID, lower, upper); err == nil {
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
		if result, err = db.Exec(`UPDATE data.latency_threshold SET lower=$3, upper=$4
				WHERE sitePK = (SELECT sitePK FROM data.site WHERE siteID = $1)
				AND typePK = (SELECT typePK FROM data.type WHERE typeID = $2)`,
			siteID, typeID, lower, upper); err == nil {
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

func dataLatencyThresholdDelete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	if _, err := db.Exec(`DELETE FROM data.latency_threshold
				WHERE sitePK = (SELECT sitePK FROM data.site WHERE siteID = $1)
				AND typePK = (SELECT typePK FROM data.type WHERE typeID = $2)`,
		v.Get("siteID"), v.Get("typeID")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func dataLatencyThresholdProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error
	var rows *sql.Rows

	v := r.URL.Query()
	typeID := v.Get("typeID")
	siteID := v.Get("siteID")

	args := []interface{}{} // empty SQL query args
	sqlQuery := `SELECT siteID, typeID, lower, upper
		FROM data.latency_threshold
		JOIN data.site USING (sitepk)
		JOIN data.type USING (typepk)`

	// Append optional arguments to sql query string and query args
	if siteID != "" && typeID != "" {
		sqlQuery += " WHERE siteID = $1 AND typeID = $2"
		args = append(args, siteID, typeID)
	} else if siteID != "" {
		sqlQuery += " WHERE siteID = $1"
		args = append(args, siteID)
	} else if typeID != "" {
		sqlQuery += " WHERE typeID = $1"
		args = append(args, typeID)
	}

	if rows, err = dbR.Query(sqlQuery, args...); err != nil {
		return weft.InternalServerError(err)
	}

	var ts mtrpb.DataLatencyThresholdResult

	for rows.Next() {
		var t mtrpb.DataLatencyThreshold

		if err = rows.Scan(&t.SiteID, &t.TypeID, &t.Lower, &t.Upper); err != nil {
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
