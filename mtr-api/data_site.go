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
	"fmt"
)

// dataSite - table data.site
// data is recorded at a site which is located at a point.
type dataSite struct{}

func (a dataSite) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "latitude", "longitude"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	siteID := v.Get("siteID")

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
	if result, err = db.Exec(`INSERT INTO data.site(siteID, latitude, longitude) VALUES($1, $2, $3)`,
		siteID, latitude, longitude); err == nil {
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
		if result, err = db.Exec(`UPDATE data.site SET latitude=$2, longitude=$3 where siteID=$1`,
			siteID, latitude, longitude); err == nil {
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

func (a dataSite) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID"}, []string{}); !res.Ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM data.site where siteID = $1`, r.URL.Query().Get("siteID")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a dataSite) allProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT siteID, latitude, longitude FROM data.site`); err != nil {
		return weft.InternalServerError(err)
	}

	var ts mtrpb.DataSiteResult

	for rows.Next() {
		var t mtrpb.DataSite

		if err = rows.Scan(&t.SiteID, &t.Latitude, &t.Longitude); err != nil {
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
