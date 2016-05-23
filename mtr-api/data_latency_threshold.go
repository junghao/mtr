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

func (a *dataLatencyThreshold) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "lower", "upper"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error

	if a.lower, err = strconv.Atoi(v.Get("lower")); err != nil {
		return weft.BadRequest("invalid lower")
	}

	if a.upper, err = strconv.Atoi(v.Get("upper")); err != nil {
		return weft.BadRequest("invalid upper")
	}

	if res := a.dataSiteType.read(v.Get("siteID"), v.Get("typeID")); !res.Ok {
		return res
	}

	// Ignore errors then update anyway.  TODO Change to upsert 9.5
	if _, err := db.Exec(`INSERT INTO data.latency_threshold(sitePK, typePK, lower, upper)
		VALUES ($1,$2,$3,$4)`,
		a.dataSite.pk, a.dataType.pk, a.lower, a.upper); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return weft.InternalServerError(err)
		}
	}

	if _, err := db.Exec(`UPDATE data.latency_threshold SET lower=$3, upper=$4
		WHERE
		sitePK = $1
		AND
		typePK = $2`,
		a.dataSite.pk, a.dataType.pk, a.lower, a.upper); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a *dataLatencyThreshold) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	if res := a.dataSiteType.read(v.Get("siteID"), v.Get("typeID")); !res.Ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM data.latency_threshold
		WHERE sitePK = $1
		AND typePK = $2 `,
		a.dataSite.pk, a.dataType.pk); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a *dataLatencyThreshold) get(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
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
