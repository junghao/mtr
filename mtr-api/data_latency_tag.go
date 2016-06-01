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

// dataLatencyTag - table data.latency_tag
type dataLatencyTag struct {
}

func (a dataLatencyTag) put(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "tag"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error
	var result sql.Result

	if result, err = db.Exec(`INSERT INTO data.latency_tag(sitePK, typePK, tagPK)
				SELECT sitePK, typePK, tagPK
				FROM data.site, data.type, mtr.tag
				WHERE siteID = $1
				AND typeID = $2
				AND tag = $3`,
		v.Get("siteID"), v.Get("typeID"), v.Get("tag")); err != nil {
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

func (a dataLatencyTag) delete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "tag"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	if _, err := db.Exec(`DELETE FROM data.latency_tag
			WHERE sitePK = (SELECT sitePK FROM data.site WHERE siteID = $1)
			AND typePK = (SELECT typePK FROM data.type WHERE typeID = $2)
			AND tagPK = (SELECT tagPK FROM mtr.tag where tag = $3)`,
		v.Get("siteID"), v.Get("typeID"), v.Get("tag")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a dataLatencyTag) all(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT siteID, tag, typeID from data.latency_tag
				JOIN mtr.tag USING (tagpk)
				JOIN data.site USING (sitepk)
				JOIN data.type USING (typepk)
				ORDER BY tag ASC`); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	var ts mtrpb.DataLatencyTagResult

	for rows.Next() {
		var t mtrpb.DataLatencyTag

		if err = rows.Scan(&t.SiteID, &t.Tag, &t.TypeID); err != nil {
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
