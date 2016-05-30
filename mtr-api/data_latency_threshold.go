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

// dataLatencyThreshold - table data.latency_threshold
type dataLatencyThreshold struct{}

func (a dataLatencyThreshold) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "lower", "upper"}, []string{}); !res.Ok {
		return res
	}

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

func (a dataLatencyThreshold) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	if _, err := db.Exec(`DELETE FROM data.latency_threshold
				WHERE sitePK = (SELECT sitePK FROM data.site WHERE siteID = $1)
				AND typePK = (SELECT typePK FROM data.type WHERE typeID = $2)`,
		v.Get("siteID"), v.Get("typeID")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a dataLatencyThreshold) get(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT siteID, typeID, lower, upper
		FROM
		data.latency_threshold JOIN data.site USING (sitepk)
		JOIN data.type USING (typepk)`); err != nil {
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

	h.Set("Content-Type", "application/x-protobuf")

	return &weft.StatusOK
}
